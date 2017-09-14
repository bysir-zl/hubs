package listener

import "github.com/bysir-zl/hubs/core/net/channel"

type Interface interface {
	Accept() (conn *channel.Channel, err error)
	Listen(addr string, isFormFd bool) (err error)
	Close() (err error)
	Fd() (fd uintptr, err error)
}
