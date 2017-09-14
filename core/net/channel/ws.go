package channel

import (
	"github.com/gorilla/websocket"
	"io"
)

type Ws struct {
	conn *websocket.Conn
}

func (p *Ws) ReadFrame() (bs []byte, err error) {
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

func (p *Ws) WriteFrame(bs []byte) (err error) {
	err = p.conn.WriteMessage(websocket.TextMessage, bs)
	return
}

func (p *Ws) Close() error {
	return p.conn.Close()
}

func FromWsConn(conn *websocket.Conn) *Channel {
	p := FromReadWriteCloser(&Ws{conn: conn})
	return p
}
