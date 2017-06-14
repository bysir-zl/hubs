package conn_wrap

type Interface interface {
	Reader() (rc chan []byte, closed bool)
	Writer() (wc chan []byte, closed bool)
	Dispatch(topic string, bs []byte, exp Interface)
	UnSubscribe(topic string) (err error)
	Subscribe(topic string) (err error)
	SetValue(key string,value interface{})
	Value(key string) (value interface{},ok bool)
	Close() (err error)
}

var BufSize = 512
