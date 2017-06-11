package conn

import (
	"github.com/gorilla/websocket"
	"io"
)

type Ws struct {
	conn *websocket.Conn
}

// 这里由于是Ws,只能一条一条的读数据,所以l无效
func (p *Ws) Read(l int64) (bs []byte, err error) {
read:
	t, bs, err := p.conn.ReadMessage()

	switch t {
	case websocket.BinaryMessage:
		return
	case websocket.CloseMessage:
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
	err = p.conn.WriteMessage(websocket.BinaryMessage, bs)
	return
}
