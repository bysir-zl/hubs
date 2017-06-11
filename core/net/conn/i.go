package conn

type Interface interface {
	Reader() (rc chan []byte)
	Writer() (wc chan []byte)
	Dispatch(topic string, bs []byte, self bool)
	UnSubscribe(topic string) (err error)
	Subscribe(topic string) (err error)
	SetValue(key string,value interface{})
	Value(key string) (value interface{},ok bool)
	Close() (err error)
}

var BufSize = 512
