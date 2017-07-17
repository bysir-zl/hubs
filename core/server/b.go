package server

import (
	"github.com/bysir-zl/hubs/core/net/listener"
	"context"
	"github.com/bysir-zl/hubs/core/net/conn_wrap"
)

type ProductNet func()
type ConnHandle func(conn_wrap.Interface)

func Run(ctx context.Context, addr string, listener listener.Interface, h ConnHandle) (err error) {
	err = listener.Listen(addr)
	if err != nil {
		return
	}

	for {
		c, e := listener.Accept(ctx)
		if e != nil {
			err = e
			return
		}
		go h(c)
	}

	return
}

func GetConnManager() *conn_wrap.Manager {
	return conn_wrap.DefManager
}
