package listener

import (
	"net/http"
	"github.com/gorilla/websocket"
	"net"
	"errors"
	"os"
	"github.com/bysir-zl/hubs/core/util"
	"github.com/bysir-zl/hubs/core/net/channel"
)

type Ws struct {
	acceptChan chan *websocket.Conn
	listener   *net.TCPListener
	isListened bool
	closeC     chan struct{}
}

const (
	readBufferSize  = 1024
	writeBufferSize = 1024
)

func (p *Ws) Accept() (c *channel.Channel, err error) {
	select {
	case tcpConn := <-p.acceptChan:
		c = channel.FromWsConn(tcpConn)
		return
	case <-p.closeC:
		err = Err_Stoped
		return
	}
	return
}

func (p *Ws) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	conn, err := websocket.Upgrade(res, req, nil, readBufferSize, writeBufferSize)
	if err != nil {
		util.Logger.Error("ws Upgrade err: ", err)
		return
	}

	p.acceptChan <- conn
}

func (p *Ws) Listen(addr string, isFromFd bool) (err error) {
	if p.isListened {
		err = errors.New("is listened")
		return
	} else {
		p.isListened = true
	}

	var l net.Listener
	if isFromFd {
		file := os.NewFile(3, "")
		l, err = net.FileListener(file)
		if err != nil {
			return
		}
	} else {
		l, err = net.Listen("tcp", addr)
		if err != nil {
			return
		}
	}
	p.listener = l.(*net.TCPListener)

	go func() {
		e := http.Serve(l, p)
		if e != nil {
			return
		}
	}()

	return
}

// 请不要close两次
func (p *Ws) Close() (err error) {
	close(p.closeC)
	err = p.listener.Close()
	return
}

func (p *Ws) Fd() (pd uintptr, err error) {
	f, err := p.listener.File()
	if err != nil {
		return
	}
	pd = f.Fd()
	return
}

func NewWs() *Ws {
	return &Ws{
		acceptChan: make(chan *websocket.Conn, 10),
		closeC:     make(chan struct{}),
	}
}
