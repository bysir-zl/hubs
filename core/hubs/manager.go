package hubs

import (
	"sync"
	"github.com/bysir-zl/hubs/core/net/channel"
)

type Manager struct {
	m map[string]map[*channel.Channel]struct{}
	sync.RWMutex
}

func (p *Manager) Subscribe(conn *channel.Channel, topic string) {
	p.Lock()
	if _, ok := p.m[topic]; ok {
		p.m[topic][conn] = struct{}{}
	} else {
		p.m[topic] = map[*channel.Channel]struct{}{conn: {}}
	}
	p.Unlock()
}

func (p *Manager) UnSubscribe(conn *channel.Channel, topic string) {
	p.Lock()
	if _, ok := p.m[topic]; ok {
		delete(p.m[topic], conn)
	}
	p.Unlock()
}

func (p *Manager) ConnByTopic(topic string) (cs []*channel.Channel) {
	p.RLock()
	if ccs, ok := p.m[topic]; ok {
		cs = make([]*channel.Channel, len(ccs))
		var i int
		for c := range ccs {
			cs[i] = c
			i++
		}
	}
	p.RUnlock()
	return
}

func (p *Manager) SendToTopic(topic string, bs []byte, expects ...*channel.Channel) (count int, err error) {
	cs := p.ConnByTopic(topic)
	if cs == nil {
		return
	}
	for _, c := range cs {
		isIn := false
		for _, exp := range expects {
			if c == exp {
				isIn = true
				break
			}
		}
		if isIn {
			continue
		}
		count++
		err = c.Write(bs)
		if err != nil {
			return
		}
	}
	return
}

func NewManager() *Manager {
	return &Manager{
		m: map[string]map[*channel.Channel]struct{}{},
	}
}
