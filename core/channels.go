package core

import "sync"

type ChannelMapper interface {
	Get(string) *ChanWithLock
	New(string, int) *ChanWithLock
	GetOrNew(string) *ChanWithLock
}

type ChanWithLock struct {
	sync.RWMutex
	c chan []byte
}

type ChannelMap struct {
	data map[string]*ChanWithLock
	size int
}

func NewChannelMap(size int) ChannelMapper {
	return &ChannelMap{
		data: map[string]*ChanWithLock{},
		size: size,
	}
}

func (p *ChannelMap) Get(k string) *ChanWithLock {
	if v, ok := p.data[k]; ok {
		return v
	}
	return nil
}

func (p *ChannelMap) New(k string, n int) *ChanWithLock {
	if p.Get(k) == nil {
		c := make(chan []byte, n)
		v := &ChanWithLock{c: c}
		p.data[k] = v
		return v
	}
	return nil
}

func (p *ChannelMap) GetOrNew(k string) *ChanWithLock {
	if p.Get(k) != nil {
		return p.Get(k)
	}
	return p.New(k, p.size)
}
