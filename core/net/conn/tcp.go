package conn

import (
	"net"
	"context"
	"log"
)

type Tcp struct {
	conn *net.TCPConn
	rc   chan []byte
	wc   chan []byte
	buf  []byte
	data map[string]interface{}
}

func (t *Tcp) SetValue(key string, value interface{}) {
	t.data[key] = value
}

func (p *Tcp) Value(key string) (value interface{}, ok bool) {
	value, ok = p.data[key]
	return
}

func (p *Tcp) Close() (err error) {
	return p.conn.Close()
}

func (p *Tcp) Reader() (rc chan []byte) {
	rc = p.rc
	return
}

func (p *Tcp) Writer() (wc chan []byte) {
	wc = p.wc
	return
}

func (p *Tcp) Dispatch(topic string, bs []byte, self bool) {
	if self {
		DefManager.SendToTopic(topic, bs, nil)
	} else {
		DefManager.SendToTopic(topic, bs, p)
	}
	return
}

func (p *Tcp) UnSubscribe(topic string) (err error) {
	DefManager.UnSubscribe(topic, p)
	return
}

func (p *Tcp) Subscribe(topic string) (err error) {
	DefManager.Subscribe(topic, p)
	return
}

func (p *Tcp) Read() (bs []byte, err error) {
	l, err := p.conn.Read(p.buf)
	if err != nil {
		return
	}
	bs = p.buf[:l]
	return
}

func (p *Tcp) Write(bs []byte) (err error) {
	_, err = p.conn.Write(bs)
	return
}

func NewTcpConn(ctx context.Context, conn *net.TCPConn) *Tcp {
	p := &Tcp{
		conn: conn,
		rc:   make(chan []byte, BufSize),
		wc:   make(chan []byte, BufSize),
		buf:  make([]byte, 1024),
		data: make(map[string]interface{}),
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
