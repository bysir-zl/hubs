package conn_wrap

import "sync"

type Base struct {
	rc   chan []byte
	wc   chan []byte
	buf  []byte
	data map[string]interface{}

	subscribeTopics map[string]struct{} // 所有注册过的Topic
	topicL          sync.RWMutex
	closed          bool
}

func (p *Base) SetValue(key string, value interface{}) {
	p.data[key] = value
}

func (p *Base) Value(key string) (value interface{}, ok bool) {
	value, ok = p.data[key]
	return
}

func (p *Base) Reader() (rc chan []byte, ok bool) {
	ok = !p.closed
	rc = p.rc
	return
}

func (p *Base) Writer() (wc chan []byte, ok bool) {
	ok = !p.closed
	wc = p.wc
	return
}

func (p *Base) Dispatch(topic string, bs []byte, exp Interface) {
	DefManager.SendToTopic(topic, bs, exp)
	return
}

func (p *Base) close() {
	p.closed = true
	return
}

func NewBase() Base {
	return Base{
		rc:              make(chan []byte, BufSize),
		wc:              make(chan []byte, BufSize),
		buf:             make([]byte, 1024),
		data:            make(map[string]interface{}),
		subscribeTopics: make(map[string]struct{}),
	}
}
