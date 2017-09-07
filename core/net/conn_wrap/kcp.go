package conn_wrap

import (
	"io"
	"github.com/bysir-zl/hubs/core/util"
	"errors"
	"github.com/bysir-zl/bygo/log"
	"net"
)

type Kcp struct {
	conn net.Conn
	Base
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
	return p.conn.Close()
}

// unused
func (p *Kcp) CloseWaitWrite() (err error) {
	if p.closed {
		err = errors.New("closed")
		return
	}
	p.closed = true

	for {
		select {
		case bs := <-p.wc:
			e := p.WriteSync(bs)
			if e != nil {
				log.Info("conn.Write Err: ", e)
			}
		default:
			goto end
		}
	}
end:

	close(p.wc)
	return p.conn.Close()
}

func (p *Kcp) ReadSync() (bs []byte, err error) {
	// 先读长度
	lBs := make([]byte, 4)
	i, err := io.ReadFull(p.conn, lBs)
	if err != nil {
		return
	}

	if i != 4 {
		err = errors.New("err header")
		return
	}
	l := util.Byte4Int32([4]byte{lBs[0], lBs[1], lBs[2], lBs[3]})
	bs = make([]byte, l)
	_, err = io.ReadFull(p.conn, bs)

	return
}

func (p *Kcp) WriteSync(bs []byte) (err error) {
	bsW := make([]byte, len(bs)+4)
	cmdB := util.Int322Byte(uint32(len(bs)))
	bsW[0] = cmdB[0]
	bsW[1] = cmdB[1]
	bsW[2] = cmdB[2]
	bsW[3] = cmdB[3]

	copy(bsW[4:], bs)
	_, err = p.conn.Write(bsW)
	return
}

func FromKcpConn(conn net.Conn) *Kcp {
	p := &Kcp{
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
