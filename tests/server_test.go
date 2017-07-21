package tests

import (
	"testing"
	"github.com/bysir-zl/hubs/core/server"
	"github.com/bysir-zl/hubs/core/net/listener"
	"log"
	"net"
	"time"
	"github.com/bysir-zl/hubs/core/net/conn_wrap"
)

func TestWsRun(t *testing.T) {
	l := listener.NewWs()
	handle := func(con conn_wrap.Interface) {
		log.Print("conn")

		con.Subscribe("room1")
		defer con.UnSubscribe("room1")

		for {
			bs, err := con.Read()
			if err != nil {
				break
			}

			log.Print(string(bs))
			con.Write([]byte("SB"))
		}

		log.Print("close")
	}
	s := server.New("127.0.0.1:10010", l, handle)
	go func() {
		time.Sleep(5 * time.Second)
		s.Stop()
	}()
	err := s.Run()
	t.Log(err)
	time.Sleep(1*time.Second)
}

func TestTcpRun(t *testing.T) {
	tcpNet := listener.NewWs()
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
	s := server.New("127.0.0.1:9900", tcpNet, handle)
	go func() {
		time.Sleep(5 * time.Second)
		s.Stop()
	}()
	s.Run()
}

func TestClient(t *testing.T) {
	a := net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9900, Zone: ""}
	c, err := net.DialTCP("tcp", nil, &a)
	if err != nil {
		t.Fatal(err)
	}

	con := conn_wrap.FromTcpConn(c)

	con.Write([]byte("hello"))

	for {
		bs, err := con.Read()
		if err != nil {
			log.Print("err ", err)
			return
		}
		log.Print("read: ", string(bs))
	}

}

func BenchmarkGg(b *testing.B) {
	a := net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9900, Zone: ""}
	c, err := net.DialTCP("tcp", nil, &a)
	if err != nil {
		b.Fatal(err)
	}
	con := conn_wrap.FromTcpConn(c)
	for i := 0; i < b.N; i++ {
		con.Write([]byte("hello"))
		{
			con.Read()
		}
	}
}
