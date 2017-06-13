package server

import (
	"time"
	"github.com/bysir-zl/hubs/exp/tank/room"
)

const UpdateTime = time.Second / 10 // 1s 同步10次

var manager = NewManager()

func Run() {
	go listen("127.0.0.1:5656")

	t := time.Tick(UpdateTime)
	preTime := time.Now()
	for {
		select {
		case <-t:
			dur := time.Now().Sub(preTime)
			preTime = time.Now()
			Update(dur)
		}
	}
}

func Update(d time.Duration) {
	manager.Update(d)
}

func init() {
	r := room.NewManager()
	manager.Add("room1", r)
}
