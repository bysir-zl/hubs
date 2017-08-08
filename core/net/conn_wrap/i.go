package conn_wrap

type Interface interface {
	Read() (bs []byte, err error)
	Write(bs []byte) (err error)
	WriteSync(bs []byte) (err error)
	SetValue(key string, value interface{})
	Value(key string) (value interface{}, ok bool)
	Close() (err error)
}

var BufSize = 512
