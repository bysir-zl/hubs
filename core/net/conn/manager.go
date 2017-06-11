package conn

import (
	"sync"
)

type Manager struct {
	m map[string]map[Interface]struct{}
	sync.RWMutex
}

var DefManager *Manager

func (p *Manager) Subscribe(topic string, conn Interface) {
	p.Lock()
	if _, ok := p.m[topic]; ok {
		p.m[topic][conn] = struct{}{}
	} else {
		p.m[topic] = map[Interface]struct{}{conn: {}}
	}
	p.Unlock()
}

func (p *Manager) UnSubscribe(topic string, conn Interface) {
	p.Lock()
	if _, ok := p.m[topic]; ok {
		delete(p.m[topic], conn)
	}
	p.Unlock()
}

func (p *Manager) ConnByTopic(topic string) (cs []Interface) {
	p.RLock()
	if ccs, ok := p.m[topic]; ok {
		cs = make([]Interface, len(ccs))
		var i int
		for c := range ccs {
			cs[i] = c
			i++
		}
	}
	p.RUnlock()
	return
}

func (p *Manager) SendToTopic(topic string, bs []byte, expect Interface) (count int, err error) {
	cs := p.ConnByTopic(topic)
	if cs == nil {
		return
	}
	for _, c := range cs {
		if c == expect {
			continue
		}
		count++
		c.Writer() <- bs
	}

	return
}

func init() {
	DefManager = &Manager{m: map[string]map[Interface]struct{}{}}
}
