package conn_wrap

import (
	"sync"
	"time"
	"sync/atomic"
	"errors"
	"bytes"
	"github.com/bysir-zl/bygo/log"
)

// 解决粘包等问题,以及心跳
type Conn struct {
	rc   chan []byte
	wc   chan []byte
	data map[string]interface{}

	subscribeTopics map[string]struct{} // 所有注册过的Topic
	topicLocker     sync.RWMutex
	closed          int32         // 原子变量
	dieC            chan struct{} // 关闭连接通知
	closeReason     error         // 关闭原因

	checkPingDuration time.Duration // 如果没收到一段时间没收到ping 则关闭连接,为0则不检查
	checkPongDuration time.Duration // 如果没收到一段时间没收到pong 则关闭连接,为0则不检查
	conn              ReadWriteCloser

	lastPingAt time.Time
	lastPongAt time.Time
}

type ReadWriteCloser interface {
	ReadFrame() ([]byte, error)
	WriteFrame([]byte) (error)
	Close() (error)
}

var (
	Ping = []byte("ping")
	Pong = []byte("pong")
)

const (
	BufSize = 256
)

func (p *Conn) SetValue(key string, value interface{}) {
	p.data[key] = value
}

func (p *Conn) Value(key string) (value interface{}, ok bool) {
	value, ok = p.data[key]
	return
}

func (p *Conn) Read() (bs []byte, err error) {
	if atomic.LoadInt32(&p.closed) > 0 {
		err = p.closeReason
		return
	}

read:
	bs, ok := <-p.rc
	if !ok {
		err = p.closeReason
		return
	}

	if bytes.Compare(bs, Ping) == 0 {
		p.lastPingAt = time.Now()
		// 回应一个pong
		p.Write(Pong)
		goto read
	} else if bytes.Compare(bs, Pong) == 0 {
		p.lastPongAt = time.Now()
		goto read
	}

	return
}

func (p *Conn) Write(bs []byte) (err error) {
	val := atomic.LoadInt32(&p.closed)
	if val > 0 {
		err = p.closeReason
		return
	}

	// 当p.wc满了阻塞的时候关闭连接 也能打破阻塞
	select {
	case <-p.dieC:
		err = p.closeReason
		return
	case p.wc <- bs:
	}
	return
}

func (p *Conn) WriteSync(bs []byte) (err error) {
	if atomic.LoadInt32(&p.closed) > 0 {
		err = p.closeReason
		return
	}

	return p.conn.WriteFrame(bs)
}

// 关闭连接, Write和Read的阻塞将会打破 并返回错误
func (p *Conn) Close(reason error) (err error) {
	if atomic.AddInt32(&p.closed, 1) > 1 {
		err = errors.New("closed")
		return
	}
	// 如果这里关闭了写通道, 在上面Write方法内如果阻塞了(p.wc满了), 就会发生 send to closed chan 的panic
	//close(p.wc)
	close(p.rc)
	close(p.dieC)

	if reason == nil {
		reason = Err_CloseDefault
	}
	p.closeReason = reason

	return p.conn.Close()
}

// 启动ping, 定时写入Ping消息
// 用于客户端
func (p *Conn) StartPing(duration time.Duration) {
	go func() {
		for range time.Tick(duration) {
			if atomic.LoadInt32(&p.closed) > 0 {
				return
			}

			p.wc <- Ping
		}
	}()
	return
}

// 检测客户端一定时间有无ping, 若没有就关闭连接, duration为0则不检查
// 用于服务端
func (p *Conn) CheckPing(duration time.Duration) {
	p.checkPingDuration = duration
	return
}

// 检测服务器一定时间有无pong, 若没有就关闭连接, duration为0则不检查
// 用于客户端
func (p *Conn) CheckPong(duration time.Duration) {
	p.checkPongDuration = duration
	return
}

func (p *Conn) monitor() {
	// 开启写协程
	go func() {
		for {
			select {
			case <-p.dieC:
				return
			case bs, ok := <-p.wc:
				if !ok {
					return
				}
				e := p.conn.WriteFrame(bs)
				if e != nil {
					log.Info("conn.Write Err: ", e)
				}
			}
		}
	}()

	temp := make(chan []byte, 8)
	// 开启读协程
	go func() {
		for {
			bs, err := p.conn.ReadFrame()
			if err != nil {
				// 有错误就直接关闭
				p.Close(errors.New("broken pipe"))
				return
			}
			temp <- bs
		}
	}()

	// 检测ping
	p.lastPingAt = time.Now()
	p.lastPongAt = time.Now()
	go func() {
		for {
			select {
			case <-p.dieC:
				return
			case <-time.Tick(1 * time.Second):
				if p.checkPingDuration != 0 && time.Now().Sub(p.lastPingAt) > p.checkPingDuration {
					// 超时没收到ping就关闭
					p.Close(Err_CloseByPing)
					return
				}
				if p.checkPongDuration != 0 && time.Now().Sub(p.lastPongAt) > p.checkPongDuration {
					// 超时没收到pong就关闭
					p.Close(Err_CloseByPong)
					return
				}
			case bs := <-temp:
				select {
				case p.rc <- bs:
				case <-p.dieC:
					return
				}
			}
		}
	}()
}

func FromReadWriteCloser(conn ReadWriteCloser) *Conn {
	return &Conn{
		rc:              make(chan []byte, BufSize),
		wc:              make(chan []byte, BufSize),
		data:            make(map[string]interface{}),
		subscribeTopics: make(map[string]struct{}),
		dieC:            make(chan struct{}),
		conn:            conn,
	}
}
