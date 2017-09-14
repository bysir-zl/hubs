package tests

import (
	"testing"
	"github.com/bysir-zl/hubs/core/net/listener"
	"log"
	"net"
	"time"
	"github.com/bysir-zl/hubs/core/hubs"
	"github.com/xtaci/kcp-go"
	"errors"
	"strconv"
	"github.com/bysir-zl/hubs/core/net/channel"
)

type Handler struct {
}

func (p *Handler) Server(s *hubs.Server, con *channel.Channel) {
	con.CheckPing(5 * time.Second)
	log.Print("conn")

	s.Subscribe(con, "room1")
	defer s.UnSubscribe(con, "room1")

	authed := false
	go func() {
		time.Sleep(5 * time.Second)
		if !authed {
			con.Close(errors.New("no auth"))
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
		con.Write([]byte("SB "+string(bs)))
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
	con := channel.FromTcpConn(c, channel.NewLenProtoCol())

	con.Write([]byte("hello11123"))
	con.Write([]byte("hello"))
	log.Print("Write")

	for {
		bs, err := con.Read()
		if err != nil {
			log.Print("err ", err)
			return
		}
		con.Write([]byte("hello"))
		log.Print("read: ", string(bs))
	}

}

func TestTcpFor(t *testing.T) {
	a := net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9900, Zone: ""}
	c, err := net.DialTCP("tcp", nil, &a)
	if err != nil {
		log.Print(err)
	}
	con := channel.FromTcpConn(c, channel.NewLenProtoCol())
	con.CheckPong(5*time.Second)
	for i := 0; i < 200; i++ {
		//con.StartPing(10 * time.Second)
		//con.CheckPong(15 * time.Second)

		con.Write([]byte("hello " + strconv.Itoa(i)))


			r,_:=con.Read()
			log.Print(string(r))
		
	}
	select {
	
	}
}

func TestKcpServer(t *testing.T) {
	tcpNet := listener.NewKcp()

	s := hubs.New(":9900", tcpNet, &Handler{})
	err := s.Run()
	t.Log(err)
}

func TestKcpClient(t *testing.T) {
	c, err := kcp.DialWithOptions(":8081", nil, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	con := channel.FromKcpConn(c, channel.NewLenProtoCol())
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

func TestKcpFor(t *testing.T) {
	c, err := kcp.DialWithOptions(":9900", nil, 10, 3)
	if err != nil {
		log.Print(err)
	}
	con := channel.FromKcpConn(c, channel.NewLenProtoCol())
	for i := 0; i < 2; i++ {
		//con.StartPing(10 * time.Second)
		//con.CheckPong(15 * time.Second)
		go func(i int) {

			con.Write([]byte("hello" + strconv.Itoa(i)))
			r, _ := con.Read()
			log.Print(string(r))
		}(i)
	}

	select {}

}

func TestClient(t *testing.T) {
	a := net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9900, Zone: ""}
	c, err := net.DialTCP("tcp", nil, &a)
	if err != nil {
		t.Fatal(err)
	}

	con := channel.FromTcpConn(c, channel.NewLenProtoCol())

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
	con := channel.FromTcpConn(c, channel.NewLenProtoCol())
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
