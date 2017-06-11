package server

import (
	"github.com/bysir-zl/hubs/core/net/listener"
	"github.com/bysir-zl/hubs/core/net/conn"
	"context"
)

type ProductNet func() listener.Interface
type ConnHandle func(conn.Interface)

func Run(ctx context.Context, addr string, p ProductNet, h ConnHandle) (err error) {
	l := p()
	err = l.Listen(addr)
	if err != nil {
		return
	}

	for {
		c, e := l.Accept(ctx)
		if e != nil {
			err = e
			return
		}
		go h(c)
	}

	return
}
