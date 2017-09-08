package listener

import (
	"github.com/bysir-zl/hubs/core/net/conn_wrap"
	"github.com/xtaci/kcp-go"
	"errors"
)

type Kcp struct {
	listener *kcp.Listener
	isClose  bool
}

func (p *Kcp) Accept() (c *conn_wrap.Conn, err error) {
	kcpConn, err := p.listener.AcceptKCP()
	if err != nil {
		if p.isClose {
			err = Err_Stoped
		}
		return
	}
	c = conn_wrap.FromKcpConn(kcpConn)
	return
}

func (p *Kcp) Listen(addr string, isFormFd bool) (err error) {
	p.listener, err = kcp.ListenWithOptions(addr,nil,0,0)
	return
}

func (p *Kcp) Close() (err error) {
	p.isClose = false
	return p.listener.Close()
}

func (p *Kcp) Fd() (pd uintptr, err error) {
	err = errors.New("kcp not suppored File")
	return
}

func NewKcp() *Kcp {
	return &Kcp{
	}
}
