package listener

import (
	"context"
	"github.com/bysir-zl/hubs/core/net/conn_wrap"
)

type Interface interface {
	Accept(ctx context.Context) (conn conn_wrap.Interface, err error)
	Listen(addr string) (err error)
	Close() (err error)
}
