package tests

import (
	"net"
	"github.com/bysir-zl/hubs/core/net/conn_wrap"
	"testing"
	"context"
	"encoding/json"
	"github.com/bysir-zl/hubs/exp/tank/server"
	"log"
	"time"
)

func TestClient(t *testing.T) {
	a := net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5656, Zone: ""}
	c, err := net.DialTCP("tcp", nil, &a)
	if err != nil {
		t.Fatal(err)
	}

	con := conn_wrap.FromTcpConn(context.Background(), c)

	auth := server.Request{
		Cmd: server.CQ_Auth,
		Body: map[string]interface{}{
			"id": 1,
		},
	}

	move := server.Request{
		Cmd: server.CQ_Move,
		Body: map[string]interface{}{
			"Ang":   1,
			"Speed": 1,
		},
	}

	room := server.Request{
		Cmd: server.CQ_GetRoom,
	}

	var bs []byte

	bs, _ = json.Marshal(auth)
	con.Writer() <- bs
	bs, _ = json.Marshal(move)
	con.Writer() <- bs

	go func() {
		for range time.Tick(time.Second / 10) {
			bs, _ = json.Marshal(room)
			con.Writer() <- bs
		}
	}()

	for r := range con.Reader() {
		log.Print(string(r))
	}

	//<-(chan int)(nil)
}

func TestN(t *testing.T) {
	var x chan int = make(chan int,10)
	x=nil
	x<-1
}