package channel

import (
	"sync"
	"time"
	"sync/atomic"
	"errors"
	"bytes"
	"github.com/bysir-zl/bygo/log"
	"github.com/bysir-zl/hubs/core/util"
)

// 解决粘包等问题,以及心跳
type Channel struct {
	rc   chan []byte
	wc   chan []byte
	data map[string]interface{}

	subscribeTopics map[string]struct{} // 所有注册过的Topic
	topicLocker     sync.RWMutex
	closed          int32         // 原子变量
	dieC            chan struct{} // 关闭连接通知
	closeReason     error         // 关闭原因

	goWrite int32
	goRead  int32

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
	BufSize = 32
)

func (p *Channel) SetValue(key string, value interface{}) {
	p.data[key] = value
}

func (p *Channel) Value(key string) (value interface{}, ok bool) {
	value, ok = p.data[key]
	return
}

func (p *Channel) Read() (bs []byte, err error) {
	if atomic.LoadInt32(&p.closed) > 0 {
		err = p.closeReason
		return
	}

	bs, ok := <-p.rc
	if !ok {
		err = p.closeReason
		return
	}

	return
}

func (p *Channel) Write(bs []byte) (err error) {
	val := atomic.LoadInt32(&p.closed)
	if val > 0 {
		err = p.closeReason
		return
	}

	// 有写的才开启协程
	if atomic.LoadInt32(&p.goWrite) == 0 {
		util.GoPool.Schedule(p.writer)
	}

	select {
	case <-p.dieC:
		err = p.closeReason
		return
	case p.wc <- bs:
	}

	return
}

func (p *Channel) WriteSync(bs []byte) (err error) {
	if atomic.LoadInt32(&p.closed) > 0 {
		err = p.closeReason
		return
	}

	return p.conn.WriteFrame(bs)
}

// 关闭连接, Write和Read的阻塞将会打破 并返回错误
func (p *Channel) Close(reason error) (err error) {
	if atomic.AddInt32(&p.closed, 1) > 1 {
		err = errors.New("closed")
		return
	}
	log.InfoT("test", "close", reason)
	// 如果这里关闭了写通道, 在上面Write方法内如果阻塞了(p.wc满了), 就会发生 send to closed chan 的panic
	close(p.dieC)
	//close(p.wc)
	//close(p.rc)

	if reason == nil {
		reason = Err_CloseDefault
	}
	p.closeReason = reason

	return p.conn.Close()
}

// 启动ping, 定时写入Ping消息
// 用于客户端
func (p *Channel) StartPing(duration time.Duration) {
	go func() {
		for range time.Tick(duration) {
			if atomic.LoadInt32(&p.closed) > 0 {
				return
			}

			p.Write(Ping)
		}
	}()
	return
}

// 检测客户端一定时间有无ping, 若没有就关闭连接, duration为0则不检查
// 用于服务端
func (p *Channel) CheckPing(duration time.Duration) {
	p.checkPingDuration = duration
	return
}

// 检测服务器一定时间有无pong, 若没有就关闭连接, duration为0则不检查
// 用于客户端
func (p *Channel) CheckPong(duration time.Duration) {
	p.checkPongDuration = duration
	return
}

func (p *Channel) reader() {
	defer func() {
		log.InfoT("test", "reader go closed")
	}()

	temp := make(chan []byte)
	// 开启读协程
	util.GoPool.Schedule(func() {
		for {
			bs, err := p.conn.ReadFrame()
			if err != nil {
				// 有错误就直接关闭
				p.Close(err)
				return
			}
			select {
			case <-p.dieC:
			case temp <- bs:
			}
		}
	})

	// 检测ping
	p.lastPingAt = time.Now()
	p.lastPongAt = time.Now()
	timer := time.After(1 * time.Second)
	for {
		select {
		case <-p.dieC:
			return
		case bs := <-temp:
			if bytes.Compare(bs, Ping) == 0 {
				p.lastPingAt = time.Now()
				// 回应一个pong
				p.Write(Pong)
			} else if bytes.Compare(bs, Pong) == 0 {
				p.lastPongAt = time.Now()
			} else {
				select {
				case <-p.dieC:
				case p.rc <- bs:
				case <-time.After(1 * time.Second):
				}
			}
			timer = time.After(1 * time.Second)

		case <-timer:
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
		}
	}

}

// 写协程
func (p *Channel) writer() {
	atomic.StoreInt32(&p.goWrite, 1)

	// 一秒之内关闭这个协程
	timer := time.After(1 * time.Second)
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

			timer = time.After(1 * time.Second)
		case <-timer:
			atomic.StoreInt32(&p.goWrite, 0)
			return
		}
	}
}

func (p *Channel) monitor() {
	p.reader()
}

// 从连接得到有个新管道, 如果没有可用的协程服务连接则会阻塞
func FromReadWriteCloser(conn ReadWriteCloser) *Channel {
	c := &Channel{
		rc:              make(chan []byte, BufSize),
		wc:              make(chan []byte, BufSize),
		data:            make(map[string]interface{}),
		subscribeTopics: make(map[string]struct{}),
		dieC:            make(chan struct{}),
		conn:            conn,
	}
	util.GoPool.Schedule(c.monitor)
	return c
}
