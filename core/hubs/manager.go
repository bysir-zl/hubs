package hubs

import (
	"sync"
	"github.com/bysir-zl/hubs/core/net/conn_wrap"
)

type Manager struct {
	m map[string]map[conn_wrap.Interface]struct{}
	sync.RWMutex
}

func (p *Manager) Subscribe(conn conn_wrap.Interface, topic string) {
	p.Lock()
	if _, ok := p.m[topic]; ok {
		p.m[topic][conn] = struct{}{}
	} else {
		p.m[topic] = map[conn_wrap.Interface]struct{}{conn: {}}
	}
	p.Unlock()
}

func (p *Manager) UnSubscribe(conn conn_wrap.Interface, topic string) {
	p.Lock()
	if _, ok := p.m[topic]; ok {
		delete(p.m[topic], conn)
	}
	p.Unlock()
}

func (p *Manager) ConnByTopic(topic string) (cs []conn_wrap.Interface) {
	p.RLock()
	if ccs, ok := p.m[topic]; ok {
		cs = make([]conn_wrap.Interface, len(ccs))
		var i int
		for c := range ccs {
			cs[i] = c
			i++
		}
	}
	p.RUnlock()
	return
}

func (p *Manager) SendToTopic(topic string, bs []byte, expects ...conn_wrap.Interface) (count int, err error) {
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
		m: map[string]map[conn_wrap.Interface]struct{}{},
	}
}
