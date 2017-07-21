package server

import (
	"github.com/bysir-zl/hubs/core/net/listener"
	"context"
	"github.com/bysir-zl/hubs/core/server"
	"log"
	"github.com/bysir-zl/hubs/core/net/conn_wrap"
	"encoding/json"
	"fmt"
	"github.com/bysir-zl/hubs/exp/tank/item"
)

func listen(addr string) (err error) {
	tcpP := func() listener.Interface {
		return listener.NewWs()
	}
	handle := func(con conn_wrap.Interface) {
		log.Print("conn")
		rc, _ := con.Reader()
		defer con.Close()

		for {
			select {
			case bs, ok := <-rc:
				if !ok {
					goto end
				}

				req := Request{}
				err := json.Unmarshal(bs, &req)
				if err != nil {
					log.Print(err)
					break
				}

				handleRequest(con, req)
			}
		}
	end:
		id := 0
		idI, ok := con.Value("id")
		if ok {
			id = idI.(int)
			room := manager.Room("room1")
			tankKey := fmt.Sprintf("tank%d", id)
			room.Del(tankKey)
		}

		log.Print("close")
	}
	ctx := context.Background()
	log.Print("running")
	err = hubs.Run(ctx, addr, tcpP, handle)
	if err != nil {
		panic(err)
	}
	return
}

func handleRequest(conn conn_wrap.Interface, request Request) {
	id := 0
	idI, ok := conn.Value("id")
	if ok {
		id = idI.(int)
	}

	log.Print(id, request)

	roomId := "room1"
	switch request.Cmd {
	case CQ_Auth:
		conn.Subscribe("room1")

		m := request.Body.(map[string]interface{})
		id := int(m["id"].(float64))
		conn.SetValue("id", id)
		conn.SetValue("roomId", roomId)
		room := manager.Room(roomId)
		tankKey := fmt.Sprintf("tank%d", id)
		if room.Item(tankKey) == nil {
			room.Add(tankKey, item.NewTank(0, 0, 0))
		}

	case CQ_Move:
		roomId, ok := conn.Value("roomId")
		if !ok{
			return
		}
		tank, ok := manager.Room(roomId.(string)).Item(fmt.Sprintf("tank%d", id)).(*item.Tank)
		if !ok {
			return
		}
		body := request.Body.(map[string]interface{})
		tank.Ang = body["Ang"].(float64)
		tank.Speed = body["Speed"].(float64)

	case CQ_GetRoom:
		roomId, _ := conn.Value("roomId")
		room := manager.Room(roomId.(string))
		r := Response{
			Cmd:  CA_Room,
			Body: room.Items,
		}

		writeConn(conn, r)
	}
}

func writeConn(conn conn_wrap.Interface, response Response) {
	bs, _ := json.Marshal(response)
	if wc, ok := conn.Writer(); ok {
		wc <- bs
	}

}
