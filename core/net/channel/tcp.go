package channel

import (
	"net"
)

type Tcp struct {
	conn       *net.TCPConn
	protoCoder ProtoCol
}

func (p *Tcp) ReadFrame() (bs []byte, err error) {
	return p.protoCoder.Read(p.conn)
}

func (p *Tcp) WriteFrame(bs []byte) (err error) {
	_, err = p.protoCoder.Write(p.conn, bs)
	return
}

func (p *Tcp) Close() (err error) {
	return p.conn.Close()
}

func FromTcpConn(conn *net.TCPConn,protoCoder ProtoCol) *Channel {
	p := FromReadWriteCloser(&Tcp{conn: conn, protoCoder: protoCoder})

	return p
}
