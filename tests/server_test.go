package tests

import (
	"testing"
	"github.com/bysir-zl/hubs/core/server"
	"github.com/bysir-zl/hubs/core/net/listener"
	"github.com/bysir-zl/hubs/core/net/conn"
	"log"
	"net"
	"context"
	"time"
)

func TestServerRun(t *testing.T) {
	tcpNet := func() listener.Interface {
		return &listener.Tcp{}
	}
	handle := func(con conn.Interface) {
		log.Print("conn")

		con.Subscribe("room1")
		defer con.UnSubscribe("room1")

		rc := con.Reader()

		stop:= make(chan struct{})
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
	i, err := c.Write([]byte("hello"))
	log.Print(i, err)

	{
		rbs := make([]byte, 1024)
		// 读到消息就关闭
		i, _ := c.Read(rbs)
		log.Print(string(rbs[:i]))
		c.Close()
	}


	//<-(chan int)(nil)
}
