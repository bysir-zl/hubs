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

func NewBase() Base {
	return Base{
		rc:              make(chan []byte, BufSize),
		wc:              make(chan []byte, BufSize),
		buf:             make([]byte, 1024),
		data:            make(map[string]interface{}),
		subscribeTopics: make(map[string]struct{}),
	}
}
