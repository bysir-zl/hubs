package server

import (
	"github.com/bysir-zl/hubs/core/net/listener"
	"context"
	"github.com/bysir-zl/hubs/core/net/conn_wrap"
)

type ConnHandle func(conn_wrap.Interface)

func Run(ctx context.Context, addr string, net listener.Interface, h ConnHandle) (err error) {
	err = net.Listen(addr)
	if err != nil {
		return
	}

	for {
		c, e := net.Accept(ctx)
		if e != nil {
			err = e
			return
		}
		go h(c)
	}

	return
}
