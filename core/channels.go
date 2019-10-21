package core

type Channeler interface {
	GetChannel() *chan []byte
}

type ChannelMap struct {
	data map[string]*chan []byte
	size int
}

func NewChannelMap(size int) *ChannelMap {
	return &ChannelMap{
		data: map[string]*chan []byte{},
		size: size,
	}
}

func (p *ChannelMap) Get(k string) *chan []byte {
	if v, ok := p.data[k]; ok {
		return v
	}
	return nil
}
func (p *ChannelMap) New(k string, n int) *chan []byte {
	if p.Get(k) == nil {
		v := make(chan []byte, n)
		p.data[k] = &v
		return &v
	}
	return nil
}

func (p *ChannelMap) GetOrNew(k string) *chan []byte {
	if p.Get(k) != nil {
		return p.Get(k)
	}
	return p.New(k, p.size)
}
