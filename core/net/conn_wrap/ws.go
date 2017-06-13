package conn_wrap

import (
	"github.com/gorilla/websocket"
	"io"
	"context"
	"log"
)

type Ws struct {
	conn *websocket.Conn
	Base
}

func (p *Ws) Close() (err error) {
	p.topicL.RLock()
	for t := range p.subscribeTopics {
		DefManager.UnSubscribe(t, p)
	}
	p.topicL.RUnlock()

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

func (p *Ws) Read() (bs []byte, err error) {
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

func (p *Ws) Write(bs []byte) (err error) {
	err = p.conn.WriteMessage(websocket.TextMessage, bs)
	return
}

func FromWsConn(ctx context.Context, conn *websocket.Conn) *Ws {
	p := &Ws{
		conn: conn,
		Base: NewBase(),
	}
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case bs := <-p.wc:
				e := p.Write(bs)
				if e != nil {
					log.Print("conn.Wirte err: ", e)
				}
			case <-ctx.Done():
				p.conn.Close()
				goto exit
			case <-stop:
				goto exit
			}
		}
	exit:
		close(p.wc)
	}()
	go func() {
		for {
			bs, err := p.Read()

			if err != nil {
				p.conn.Close()
				goto exit
			}

			p.rc <- bs
		}
	exit:
		close(p.rc)
		close(stop)
	}()
	return p
}
