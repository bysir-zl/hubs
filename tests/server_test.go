package tests

import (
	"testing"
	"github.com/bysir-zl/hubs/core/net/listener"
	"log"
	"net"
	"time"
	"github.com/bysir-zl/hubs/core/net/conn_wrap"
	"github.com/bysir-zl/hubs/core/hubs"
	"github.com/xtaci/kcp-go"
)

type Handler struct {
}

func (p *Handler) Server(s *hubs.Server, con *conn_wrap.Conn) {
	con.CheckPing(15 * time.Second)

	log.Print("conn")

	s.Subscribe(con, "room1")
	defer s.UnSubscribe(con, "room1")

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
			log.Print("read err: ", err)
			break
		}
		authed = true
		log.Print(string(bs))
		con.Write([]byte("SB"))
	}

	log.Print("close")
}

func TestWsRun(t *testing.T) {
	l := listener.NewWs()

	s := hubs.New("127.0.0.1:10010", l, &Handler{})
	go func() {
		time.Sleep(5 * time.Second)
		s.Stop()
	}()
	err := s.Run()
	t.Log(err)
	time.Sleep(1 * time.Second)
}

func TestTcpServer(t *testing.T) {
	tcpNet := listener.NewTcp()

	s := hubs.New("127.0.0.1:9900", tcpNet, &Handler{})
	go func() {
		time.Sleep(5 * time.Second)
		//s.Stop()
	}()
	err := s.Run()
	t.Log(err)

}

func TestTcpClient(t *testing.T) {
	a := net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9900, Zone: ""}
	c, err := net.DialTCP("tcp", nil, &a)
	if err != nil {
		t.Fatal(err)
	}
	log.Print("conned")
	con := conn_wrap.FromTcpConn(c, conn_wrap.NewLenProtoCoder())

	con.Write([]byte("hello"))
	log.Print("Write")

	for {
		bs, err := con.Read()
		if err != nil {
			log.Print("err ", err)
			return
		}
		log.Print("read: ", string(bs))
	}

}

func TestKcpServer(t *testing.T) {
	tcpNet := listener.NewKcp()

	s := hubs.New(":9900", tcpNet, &Handler{})
	err := s.Run()
	t.Log(err)
}

func TestKcpClient(t *testing.T) {
	c, err := kcp.DialWithOptions(":9900", nil, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	con := conn_wrap.FromKcpConn(c)
	con.StartPing(10 * time.Second)
	con.CheckPong(15 * time.Second)

	con.Write([]byte("hello"))

	for {
		bs, err := con.Read()
		if err != nil {
			log.Print("err ", err)
			return
		}
		log.Print("read: ", string(bs))
	}

	time.Sleep(time.Second)
}

func TestClient(t *testing.T) {
	a := net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9900, Zone: ""}
	c, err := net.DialTCP("tcp", nil, &a)
	if err != nil {
		t.Fatal(err)
	}

	con := conn_wrap.FromTcpConn(c, conn_wrap.NewLenProtoCoder())

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
	con := conn_wrap.FromTcpConn(c, conn_wrap.NewLenProtoCoder())
	for i := 0; i < b.N; i++ {
		con.Write([]byte("hello"))
		{
			con.Read()
		}
	}
}

func TestChannel(t *testing.T) {
	var b = make(chan int, 10)
	b <- 1
	close(b)
	// 这里会出错
	b <- 2
}
