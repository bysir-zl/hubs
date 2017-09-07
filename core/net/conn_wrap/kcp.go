package conn_wrap

import (
	"errors"
	"github.com/bysir-zl/bygo/log"
	"github.com/xtaci/kcp-go"
	"bytes"
	"time"
)

type Kcp struct {
	conn *kcp.UDPSession
	Base
}

// 可自定义分包器
func (p *Kcp) SetProtoCoder(protoCoder ProtoCoder) {
	p.Base.protoCoder = protoCoder
}

func (p *Kcp) Read() (bs []byte, err error) {
	if p.closed {
		err = errors.New("closed")
		return
	}
	bs, ok := <-p.rc
	if !ok {
		err = errors.New("closed")
		return
	}
	return
}

func (p *Kcp) Write(bs []byte) (err error) {
	if p.closed {
		err = errors.New("closed")
		return
	}
	p.wc <- bs
	return
}

func (p *Kcp) Close() (err error) {
	if p.closed {
		err = errors.New("closed")
		return
	}
	p.closed = true

	close(p.wc)
	close(p.rc)
	return p.conn.Close()
}

func (p *Kcp) ReadSync() (bs []byte, err error) {
	return p.protoCoder.Read(p.conn)
}

func (p *Kcp) WriteSync(bs []byte) (err error) {
	_, err = p.protoCoder.Write(p.conn, bs)
	return
}

func FromKcpConn(conn *kcp.UDPSession) *Kcp {
	p := &Kcp{
		conn: conn,
		Base: NewBase(),
	}
	stop := make(chan struct{})

	// 开启写协程
	go func() {
		defer func() {
			p.Close()
			log.InfoT("test", "close")
		}()
		for {
			select {
			case <-stop:
				return
			case bs, ok := <-p.wc:
				if !ok {
					return
				}
				e := p.WriteSync(bs)
				if e != nil {
					log.Info("conn.Write Err: ", e)
				}
			}
		}
	}()

	// 开启读协程
	go func() {
		// 检测消息
		lastPingAt := time.Now()

		temp := make(chan []byte, 1000)
		for {
			bs, err := p.ReadSync()
			if err != nil {
				close(stop)
				return
			}
			temp <- bs
			for {
				select {
				case by := <-temp:
					if bytes.Compare(by, Ping) {
						lastPingAt = time.Now()
					} else {
						p.rc <- bs
					}
					break
				case time.Tick(1 * time.Second):
					if time.Now().Sub(lastPingAt) > p.checkPingDuration {
						close(stop)
						return
					}
				}
			}
		}
	}()

	return p
}
