package core

import "sync"

var HUB_MAP = HubMap{}

type HubMap struct {
	sync.RWMutex
	maps map[string]*Hub
}

func (m *HubMap) GetHub(id string) *Hub {
	m.Lock()
	defer m.Unlock()
	if v, ok := m.maps[id]; ok {
		return v
	} else {
		rv := NewHub()
		m.maps[id] = rv
		return rv
	}
}
