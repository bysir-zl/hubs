package conn_wrap

import (
	"sync"
	"time"
)

type Base struct {
	rc   chan []byte
	wc   chan []byte
	buf  []byte
	data map[string]interface{}

	subscribeTopics map[string]struct{} // 所有注册过的Topic
	topicLocker     sync.RWMutex
	closed          bool

	protoCoder ProtoCoder

	checkPingDuration time.Duration // 为0则不检查
}

var (
	Ping = []byte("ping")
	Pong = []byte("pong")
)

func (p *Base) SetValue(key string, value interface{}) {
	p.data[key] = value
}

func (p *Base) Value(key string) (value interface{}, ok bool) {
	value, ok = p.data[key]
	return
}

// 启动ping, 定时写入Ping消息
// 用于客户端
func (p *Base) StartPing(duration time.Duration) {
	go func() {
		for range time.Tick(duration) {
			if p.closed {
				return
			}

			p.wc <- Ping
		}
	}()
	return
}

// 检测客户端一定时间有无ping, 若没有就关闭连接, duration为0则不检查
// 用于服务端
func (p *Base) CheckPing(duration time.Duration) {
	p.checkPingDuration = duration
	return
}

func NewBase() Base {
	return Base{
		rc:              make(chan []byte, BufSize),
		wc:              make(chan []byte, BufSize),
		buf:             make([]byte, 1024),
		data:            make(map[string]interface{}),
		subscribeTopics: make(map[string]struct{}),
		protoCoder:      NewLenProtocal(),
	}
}
