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

func TestRun(t *testing.T) {
	tcpNet := &listener.Tcp{}
	handle := func(con conn_wrap.Interface) {
		log.Print("conn")

		con.Subscribe("room1")
		defer con.UnSubscribe("room1")

		authed := false
		go func() {
			time.Sleep(5 * time.Second)
			if !authed {
				con.Close()
			}
		}()

		for {
			bs, err := con.Read()
			if err != nil {
				break
			}
			authed = true
			log.Print(string(bs))
			con.Write([]byte("SB"))
		}

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

	con.Write([]byte("hello"))

	for {
		bs, err := con.Read()
		if err != nil {
			log.Print("err ", err)
			return
		}
		log.Print("read: ",string(bs))
	}

}

func BenchmarkGg(b *testing.B) {
	a := net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9900, Zone: ""}
	c, err := net.DialTCP("tcp", nil, &a)
	if err != nil {
		b.Fatal(err)
	}
	con := conn_wrap.FromTcpConn(context.Background(), c)
	for i := 0; i < b.N; i++ {
		con.Write([]byte("hello"))
		{
			con.Read()
		}
	}
}
