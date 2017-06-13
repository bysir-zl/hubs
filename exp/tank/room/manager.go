package room

import (
	"sync"
	"github.com/bysir-zl/hubs/exp/tank/item"
	"time"
	"encoding/json"
)

type Manager struct {
	Items map[string]item.Interface
	itemL sync.RWMutex
	bs []byte
}

func (p *Manager) Add(id string, item item.Interface) {
	p.itemL.Lock()
	p.Items[id] = item
	p.itemL.Unlock()
}

func (p *Manager) Del(id string) {
	p.itemL.Lock()
	delete(p.Items, id)
	p.itemL.Unlock()
}

func (p *Manager) Item(id string) (i item.Interface) {
	p.itemL.RLock()
	i = p.Items[id]
	p.itemL.RUnlock()
	return
}

// 更新一个房子里的所有物体
func (p *Manager) Update(d time.Duration) {
	p.itemL.RLock()
	for _, i := range p.Items {
		i.Update(d)
	}
	p.bs , _ = json.Marshal(p.Items)

	p.itemL.RUnlock()
}

func (p *Manager) Marshal() (bs []byte) {
	return p.bs
}

func (p *Manager) UnMarshal(bs []byte) (err error) {
	err = json.Unmarshal(bs, p.Items)
	return
}

func NewManager() *Manager {
	return &Manager{
		Items: map[string]item.Interface{},
	}
}
