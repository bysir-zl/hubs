package server

import (
	"sync"
	"time"
	"github.com/bysir-zl/hubs/exp/tank/room"
)

type Manager struct {
	items map[string]*room.Manager
	sync.RWMutex
}

func (p *Manager) Add(roomId string, item *room.Manager) {
	p.Lock()
	p.items[roomId] = item
	p.Unlock()
}

func (p *Manager) Room(roomId string) (item *room.Manager) {
	p.RLock()
	item = p.items[roomId]
	p.RUnlock()
	return
}

func (p *Manager) Del(roomId string) {
	p.Lock()
	delete(p.items, roomId)
	p.Unlock()
}

// 更新一个服务器所有房间
func (p *Manager) Update(d time.Duration) {
	p.RLock()
	for _, i := range p.items {
		go i.Update(d)
	}
	p.RUnlock()
}

func NewManager() *Manager {
	return &Manager{
		items: map[string]*room.Manager{},
	}
}
