package conn_wrap

import (
	"net"
	"context"
	"log"
	"io"
	"github.com/bysir-zl/hubs/core/util"
	"errors"
)

type Tcp struct {
	conn *net.TCPConn
	Base
}

func (p *Tcp) Close() (err error) {
	p.close()

	p.topicL.RLock()
	for t := range p.subscribeTopics {
		DefManager.UnSubscribe(t, p)
	}
	p.topicL.RUnlock()

	return p.conn.Close()
}

func (p *Tcp) UnSubscribe(topic string) (err error) {
	p.topicL.Lock()
	delete(p.subscribeTopics, topic)
	p.topicL.Unlock()

	DefManager.UnSubscribe(topic, p)
	return
}

func (p *Tcp) Subscribe(topic string) (err error) {
	p.topicL.Lock()
	p.subscribeTopics[topic] = struct{}{}
	p.topicL.Unlock()

	DefManager.Subscribe(topic, p)
	return
}

func (p *Tcp) Read() (bs []byte, err error) {
	// 先读长度
	lBs := make([]byte, 4)
	i, err := io.ReadFull(p.conn, lBs)
	if err != nil {
		return
	}

	if i != 4 {
		err = errors.New("err header")
		return
	}
	l := util.Byte4Int32([4]byte{lBs[0], lBs[1], lBs[2], lBs[3]})
	bs = make([]byte, l)
	_, err = io.ReadFull(p.conn, bs)

	return
}

func (p *Tcp) Write(bs []byte) (err error) {
	bsW := make([]byte, len(bs)+4)
	cmdB := util.Int322Byte(uint32(len(bs)))
	bsW[0] = cmdB[0]
	bsW[1] = cmdB[1]
	bsW[2] = cmdB[2]
	bsW[3] = cmdB[3]

	copy(bsW[4:], bs)
	_, err = p.conn.Write(bsW)
	return
}

func FromTcpConn(ctx context.Context, conn *net.TCPConn) *Tcp {
	p := &Tcp{
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
		p.close()
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
		p.close()
		close(p.rc)
		close(stop)
	}()
	return p
}
