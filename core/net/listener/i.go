package listener

import (
	"github.com/bysir-zl/hubs/core/net/conn"
	"context"
)

type Interface interface {
	Accept(ctx context.Context) (conn conn.Interface, err error)
	Listen(addr string) (err error)
	Close() (err error)
}
