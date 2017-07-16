package tests

import (
	"testing"
	"github.com/bysir-zl/hubs/core/server"
	"github.com/bysir-zl/hubs/core/net/listener"
	"log"
	"net"
	"context"
	"time"
	"github.com/bysir-zl/hubs/core/net/conn_wrap"
)

func TestServerRun(t *testing.T) {
	tcpNet := &listener.Tcp{}
	handle := func(con conn_wrap.Interface) {
		log.Print("conn")

		con.Subscribe("room1")
		defer con.UnSubscribe("room1")

		rc := con.Reader()

		stop := make(chan struct{})
		authed := false
		go func() {
			time.Sleep(5 * time.Second)
			if !authed {
				close(stop)
			}
		}()

		for {
			select {
			case bs, ok := <-rc:
				if !ok {
					goto end
				}
				authed = true
				log.Print(string(bs))
				con.Writer() <- []byte("SB")
				con.Writer() <- []byte("SB")
				con.Writer() <- []byte("SB")
				con.Writer() <- []byte("SB")
				con.Writer() <- []byte("SB")
				con.Writer() <- []byte("SB")
				con.Writer() <- []byte("SB")
			case <-stop:
				log.Print("auth timeout")
				con.Close()
				goto end
			}
		}
	end:
		log.Print("close")
	}
	ctx := context.Background()
	server.Run(ctx, "127.0.0.1:9900", tcpNet, handle)
}

func TestClient(t *testing.T) {
	a := net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9900, Zone: ""}
	c, err := net.DialTCP("tcp", nil, &a)
	if err != nil {
		t.Fatal(err)
	}

	con := conn_wrap.FromTcpConn(context.Background(), c)

	con.Writer() <- []byte("hello")

	for r := range con.Reader() {
		log.Print(string(r))
	}

	//<-(chan int)(nil)
}

func BenchmarkGg(b *testing.B) {
	a := net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9900, Zone: ""}
	c, err := net.DialTCP("tcp", nil, &a)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		c.Write([]byte("hello"))
		rbs := make([]byte, 1024)

		{
			c.Read(rbs)
		}
	}
}
