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
	goCheck int32

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

	for {
		bs, err = p.conn.ReadFrame()
		if err != nil {
			// 有错误就直接关闭
			p.Close(err)
			return
		}
		if bytes.Compare(bs, Ping) == 0 {
			p.lastPingAt = time.Now()
			// 回应一个pong
			p.Write(Pong)
			continue
		} else if bytes.Compare(bs, Pong) == 0 {
			p.lastPongAt = time.Now()
			continue
		}

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
	if atomic.CompareAndSwapInt32(&p.goWrite, 0, 1) {
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
	if atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		err = errors.New("closed")
		return
	}

	close(p.dieC)
	// 如果这里关闭了写通道, 就会发生 send to closed chan 的panic
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
	if atomic.CompareAndSwapInt32(&p.goCheck, 0, 1) {
		util.GoPool.Schedule(p.checker)
	}
	p.checkPingDuration = duration
	return
}

// 检测服务器一定时间有无pong, 若没有就关闭连接, duration为0则不检查
// 用于客户端
func (p *Channel) CheckPong(duration time.Duration) {
	if atomic.CompareAndSwapInt32(&p.goCheck, 0, 1) {
		util.GoPool.Schedule(p.checker)
	}
	p.checkPongDuration = duration
	return
}

// 检测ping
func (p *Channel) checker() {
	defer func() {
		log.InfoT("test", "checker go closed")
	}()

	p.lastPingAt = time.Now()
	p.lastPongAt = time.Now()
	// 检测ping
	timer := time.After(1 * time.Second)
	for {
		select {
		case <-p.dieC:
			return
		case <-timer:
			if p.checkPingDuration == 0 && p.checkPongDuration == 0 {
				atomic.StoreInt32(&p.goCheck, 0)
				return
			}

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
			timer = time.After(1 * time.Second)
		}
	}

}

// 写协程
func (p *Channel) writer() {
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

// 启动
func (p *Channel) monitor() {
}

// 从连接得到有个新管道, 如果没有可用的协程服务连接则会阻塞
func FromReadWriteCloser(conn ReadWriteCloser) *Channel {
	c := &Channel{
		rc:              make(chan []byte),
		wc:              make(chan []byte, BufSize),
		data:            make(map[string]interface{}),
		subscribeTopics: make(map[string]struct{}),
		dieC:            make(chan struct{}),
		conn:            conn,
	}
	c.monitor()
	return c
}
