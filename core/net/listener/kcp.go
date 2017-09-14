package listener

import (
	"github.com/xtaci/kcp-go"
	"errors"
	"github.com/bysir-zl/hubs/core/net/channel"
)

type Kcp struct {
	listener   *kcp.Listener
	isClose    bool
	protoCoder channel.ProtoCol
}

func (p *Kcp) Accept() (c *channel.Channel, err error) {
	kcpConn, err := p.listener.AcceptKCP()
	if err != nil {
		if p.isClose {
			err = Err_Stoped
		}
		return
	}
	c = channel.FromKcpConn(kcpConn, p.protoCoder)
	return
}

func (p *Kcp) Listen(addr string, isFormFd bool) (err error) {
	p.listener, err = kcp.ListenWithOptions(addr, nil, 10, 3)
	return
}

func (p *Kcp) Close() (err error) {
	p.isClose = false
	return p.listener.Close()
}

func (p *Kcp) Fd() (pd uintptr, err error) {
	err = errors.New("kcp not supported File")
	return
}

func NewKcp() *Kcp {
	return &Kcp{
		protoCoder: channel.NewLenProtoCol(),
	}
}
