package listener

import (
	"net"
	"strings"
	"strconv"
	"github.com/bysir-zl/hubs/core/net/conn_wrap"
)

type Tcp struct {
	listener   *net.TCPListener
	isClose    bool
	protoCoder conn_wrap.ProtoCoder
}

func (p *Tcp) Accept() (c *conn_wrap.Conn, err error) {
	tcpConn, err := p.listener.AcceptTCP()
	if err != nil {
		if p.isClose {
			err = Err_Stoped
		}
		return
	}
	c = conn_wrap.FromTcpConn(tcpConn, p.protoCoder)
	return
}

func (p *Tcp) Listen(addr string, isFormFd bool) (err error) {
	ipP := strings.Split(addr, ":")
	port, _ := strconv.Atoi(ipP[1])
	a := net.TCPAddr{IP: net.ParseIP(ipP[0]), Port: port}
	p.listener, err = net.ListenTCP("tcp", &a)
	return
}

func (p *Tcp) Close() (err error) {
	p.isClose = false
	return p.listener.Close()
}

func (p *Tcp) Fd() (pd uintptr, err error) {
	f, err := p.listener.File()
	if err != nil {
		return
	}
	pd = f.Fd()
	return
}

func NewTcp() *Tcp {
	return &Tcp{
		protoCoder: conn_wrap.NewLenProtoCoder(),
	}
}
