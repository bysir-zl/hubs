package listener

import (
	"github.com/bysir-zl/hubs/core/net/conn"
	"net"
	"strings"
	"strconv"
	"context"
)

type Tcp struct {
	listener *net.TCPListener
}

func (p *Tcp) Accept(ctx context.Context) (c conn.Interface, err error) {
	err = ctx.Err()
	if err != nil {
		return
	}
	tcpConn, err := p.listener.AcceptTCP()
	if err != nil {
		return
	}
	c = conn.NewTcpConn(ctx, tcpConn)
	return
}

func (p *Tcp) Listen(addr string) (err error) {
	ipP := strings.Split(addr, ":")
	port, _ := strconv.Atoi(ipP[1])
	a := net.TCPAddr{IP: net.ParseIP(ipP[0]), Port: port}
	p.listener, err = net.ListenTCP("tcp", &a)
	return
}

func (p *Tcp) Close() (err error) {
	return p.listener.Close()
}
