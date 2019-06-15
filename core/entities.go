package core

import (
	"net/http"
	"sync"
	"time"
)

var HUB = &Hub{}

type MessageType string

const (
	MessageTypePlain MessageType = "PLAIN"
	MessageTypeHTML  MessageType = "HTML"
	MessageTypeImage MessageType = "IMAGE"
)

type Message struct {
	Type      MessageType
	Data      []byte
	SourceReq *http.Request
	SourceWS  *WebSocket
}

func (p *Message) Str() string {
	if In(p.Type, MessageTypePlain, MessageTypeHTML) {
		return string(p.Data)
	}
	return ""
}

type Topic struct {
	sync.RWMutex
	Topic     string
	Subs      map[string]*WebSocket
	Pubs      map[string]*WebSocket
	CreatedAt time.Time
	UpdatedAt time.Time
	Close     chan (bool)
}

func (p *Topic) Sub(ws *WebSocket) {
	p.Lock()
	defer p.Unlock()
	if _, ok := p.Subs[ws.key]; !ok {
		p.Subs[ws.key] = ws
	}
}

func (p *Topic) Pub(msg *Message) {
	p.Lock()
	defer p.Unlock()

	if msg.SourceWS != nil {
		ws := msg.SourceWS
		if _, ok := p.Pubs[ws.key]; !ok {
			p.Pubs[ws.key] = ws
		}
	}

	for _, sub := range p.Subs {
		go sub.Send(msg)
	}
}

type Hub struct {
	sync.Mutex
	Topics map[string]*Topic
}

func (p *Hub) GetTopic(topic string) *Topic {
	p.Lock()
	defer p.Unlock()

	if tpc, ok := p.Topics[topic]; ok {
		return tpc
	}
	rv := &Topic{Topic: topic}
	p.Topics[topic] = rv
	return rv
}

func (p *Hub) Sub(topic string, ws *WebSocket) {
	tpc := p.GetTopic(topic)
	tpc.Sub(ws)
}

func (p *Hub) Pub(topic string, msg *Message) {
	tpc := p.GetTopic(topic)
	tpc.Pub(msg)
}
