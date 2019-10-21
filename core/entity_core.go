package core

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

var HUB = &Hub{
	Topics: map[string]*Topic{},
}

const GlobalTopicID = "global"

// representation of internal messages
const (
	MTPlain    string = "PLAIN"
	MTMarkdown string = "MARKDOWN"
	MTJSON     string = "JSON"
	MTHTML     string = "HTML"
	MTImage    string = "IMAGE"
)

var MTAll = []string{MTPlain, MTMarkdown, MTJSON, MTHTML, MTImage}

// representation of websocket messages
const (
	MTFeedback string = "FEEDBACK" // used for async event feedback
	MTResponse string = "RESPONSE" // used for message response
	MTMessage  string = "MESSAGE"  // used for publish messages
)

// http client message
type Message struct {
	Type      string `json:"type"`
	Data      string `json:"data"` // string or base64 of bytes
	SourceReq *http.Request
	SourceWS  *WebSocket
}

type Topic struct {
	sync.RWMutex
	Topic     string                `json:"topic"`
	Subs      map[string]*WebSocket `json:"subs"`
	Pubs      map[string]*WebSocket `json:"pubs"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	Close     chan (bool)           `json:"-"`
}

func (p *Topic) Sub(ws *WebSocket) {
	p.Lock()
	defer p.Unlock()
	if _, ok := p.Subs[ws.Key]; !ok {
		p.Subs[ws.Key] = ws
		p.UpdatedAt = time.Now()
	}
}

func (p *Topic) Pub(msg *Message) {
	p.Lock()
	defer p.Unlock()

	if msg.SourceWS != nil {
		ws := msg.SourceWS
		if _, ok := p.Pubs[ws.Key]; !ok {
			p.Pubs[ws.Key] = ws
			p.UpdatedAt = time.Now()
		}
	}

	// save into in-memoery buffers
	success := BufPub(p.Topic, ToJSON(msg))
	log.Printf("buffered on topic %v, %v %v\n", p.Topic, success, string(ToJSON(msg)))

	c := 0
	for _, sub := range p.Subs {
		// do not send back to self
		if sub != msg.SourceWS {
			go sub.send(p.Topic, msg)
			c++
		}
	}
	if msg.SourceWS != nil {
		msg.SourceWS.WriteSafe(ToJSON(map[string]string{
			"type":    "FEEDBACK",
			"message": fmt.Sprintf(`sent to total %v subscribers on topic "%s"`, c, p.Topic),
		}))
	}
}

func (p *Topic) removeWS(ws *WebSocket) {
	delete(p.Subs, ws.Key)
	delete(p.Pubs, ws.Key)
}

type Hub struct {
	sync.Mutex
	Topics map[string]*Topic `json:"topics"`
}

func (p *Hub) GetTopic(topic string) *Topic {
	p.Lock()
	defer p.Unlock()

	if tpc, ok := p.Topics[topic]; ok {
		return tpc
	}
	rv := &Topic{
		Topic:     topic,
		Subs:      map[string]*WebSocket{},
		Pubs:      map[string]*WebSocket{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Close:     make(chan bool, 1),
	}
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

func (p *Hub) removeWS(ws *WebSocket) {
	for _, tpc := range p.Topics {
		tpc.removeWS(ws)
	}
}
