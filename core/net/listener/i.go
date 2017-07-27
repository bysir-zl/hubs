package listener

import (
	"github.com/bysir-zl/hubs/core/net/conn_wrap"
)

type Interface interface {
	Accept() (conn conn_wrap.Interface, err error)
	Listen(addr string,isFormFd bool) (err error)
	Close() (err error)
	Fd() (fd uintptr, err error)
}
