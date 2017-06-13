package listener

import (
	"context"
	"github.com/bysir-zl/hubs/core/net/conn_wrap"
	"net/http"
	"github.com/gorilla/websocket"
	"log"
)

type Ws struct {
	acceptChan chan *websocket.Conn
}

const (
	readBufferSize  = 1024
	writeBufferSize = 1024
)

func (p *Ws) Accept(ctx context.Context) (c conn_wrap.Interface, err error) {
	err = ctx.Err()
	if err != nil {
		return
	}
	tcpConn := <-p.acceptChan
	c = conn_wrap.FromWsConn(ctx, tcpConn)
	return
}

func (p *Ws) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	conn, err := websocket.Upgrade(res, req, nil, readBufferSize, writeBufferSize)
	if err != nil {
		log.Print(err)
		return
	}
	p.acceptChan <- conn
}

func (p *Ws) Listen(addr string) (err error) {
	http.Handle("/ws", p)
	go func() {
		e := http.ListenAndServe(addr, nil)
		if e != nil {
			panic(e)
			return
		}
	}()
	return
}

func (p *Ws) Close() (err error) {
	return
}

func NewWs() *Ws {
	return &Ws{
		acceptChan: make(chan *websocket.Conn, 10),
	}
}
