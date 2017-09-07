package conn_wrap

import (
	"net"
	"errors"
	"github.com/bysir-zl/bygo/log"
)

type Tcp struct {
	conn *net.TCPConn
	Base
}

// 可自定义分包器, 默认是lengthfield
func (p *Tcp) SetProtoCoder(protoCoder ProtoCoder) {
	p.Base.protoCoder = protoCoder
}

func (p *Tcp) Read() (bs []byte, err error) {
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

func (p *Tcp) Write(bs []byte) (err error) {
	if p.closed {
		err = errors.New("closed")
		return
	}
	p.wc <- bs
	return
}

func (p *Tcp) Close() (err error) {
	if p.closed {
		err = errors.New("closed")
		return
	}
	p.closed = true

	close(p.wc)
	close(p.rc)
	return p.conn.Close()
}

func (p *Tcp) ReadSync() (bs []byte, err error) {
	return p.protoCoder.Read(p.conn)
}

func (p *Tcp) WriteSync(bs []byte) (err error) {
	_, err = p.protoCoder.Write(p.conn, bs)
	return
}

func FromTcpConn(conn *net.TCPConn) *Tcp {
	p := &Tcp{
		conn: conn,
		Base: NewBase(),
	}
	stop := make(chan struct{})

	// 开启写协程
	go func() {
		defer func() {
			p.Close()
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
		for {
			bs, err := p.ReadSync()
			if err != nil {
				close(stop)
				return
			}
			p.rc <- bs
		}
	}()

	return p
}
