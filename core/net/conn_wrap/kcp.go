package conn_wrap

import (
	"github.com/xtaci/kcp-go"
)

type Kcp struct {
	conn       *kcp.UDPSession
	protoCoder ProtoCoder
}

func (p *Kcp) ReadFrame() (bs []byte, err error) {
	return p.protoCoder.Read(p.conn)
}

func (p *Kcp) WriteFrame(bs []byte) (err error) {
	_, err = p.protoCoder.Write(p.conn, bs)
	return
}
func (p *Kcp) Close() (err error) {
	return p.conn.Close()
}

func FromKcpConn(conn *kcp.UDPSession) *Conn {
	p := FromReadWriteCloser(&Kcp{conn: conn, protoCoder: NewLenProtoCoder()})
	p.monitor()
	return p
}
