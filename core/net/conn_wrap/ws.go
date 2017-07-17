package conn_wrap

import (
	"github.com/gorilla/websocket"
	"io"
	"context"
	"log"
	"errors"
)

type Ws struct {
	conn *websocket.Conn
	Base
}

func (p *Ws) Read() (bs []byte, err error) {
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

func (p *Ws) Write(bs []byte) (err error) {
	if p.closed {
		err = errors.New("closed")
		return
	}
	p.wc <- bs
	return
}

func (p *Ws) Close() (err error) {
	if p.closed {
		err = errors.New("closed")
		return
	}
	p.close()

	p.topicL.RLock()
	for t := range p.subscribeTopics {
		DefManager.UnSubscribe(t, p)
	}
	p.topicL.RUnlock()

	close(p.wc)
	close(p.rc)
	return p.conn.Close()
}

func (p *Ws) UnSubscribe(topic string) (err error) {
	p.topicL.Lock()
	delete(p.subscribeTopics, topic)
	p.topicL.Unlock()

	DefManager.UnSubscribe(topic, p)
	return
}

func (p *Ws) Subscribe(topic string) (err error) {
	p.topicL.Lock()
	p.subscribeTopics[topic] = struct{}{}
	p.topicL.Unlock()

	DefManager.Subscribe(topic, p)
	return
}

func (p *Ws) ReadSync() (bs []byte, err error) {
read:
	t, bs, err := p.conn.ReadMessage()
	switch t {
	case websocket.BinaryMessage:
		return
	case websocket.TextMessage:
		return
	case websocket.CloseMessage, -1:
		err = io.EOF
		return
	default:
		// 忽略这条消息
		// 比如ping pong应该自己实现,和统一tcp
		goto read
	}
	return
}

func (p *Ws) WriteSync(bs []byte) (err error) {
	err = p.conn.WriteMessage(websocket.TextMessage, bs)
	return
}

func FromWsConn(ctx context.Context, conn *websocket.Conn) *Ws {
	p := &Ws{
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
			case <-ctx.Done():
				return
			case bs ,ok:= <-p.wc:
				if !ok {
					return
				}
				e := p.WriteSync(bs)
				if e != nil {
					log.Print("conn.Wirte err: ", e)
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
