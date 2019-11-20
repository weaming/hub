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

func (t *Topic) Sub(ws *WebSocket) {
	t.Lock()
	defer t.Unlock()
	if _, ok := t.Subs[ws.Key]; !ok {
		t.Subs[ws.Key] = ws
		t.UpdatedAt = time.Now()
	}
}

func (t *Topic) Pub(msg *Message) {
	t.Lock()
	defer t.Unlock()

	if msg.SourceWS != nil {
		ws := msg.SourceWS
		if _, ok := t.Pubs[ws.Key]; !ok {
			t.Pubs[ws.Key] = ws
			t.UpdatedAt = time.Now()
		}
	}

	// save into in-memoery buffers
	success := BufPub(t.Topic, ToJSON(msg))
	log.Printf("buffered on topic %v, %v %v\n", t.Topic, success, string(ToJSON(msg)))

	c := 0
	for _, sub := range t.Subs {
		// do not send back to self
		if sub != msg.SourceWS {
			go sub.send(t.Topic, msg)
			c++
		}
	}
	if msg.SourceWS != nil {
		msg.SourceWS.WriteSafe(ToJSON(map[string]string{
			"type":    "FEEDBACK",
			"message": fmt.Sprintf(`sent to total %v subscribers on topic "%s"`, c, t.Topic),
		}))
	}
}

func (t *Topic) dereferenceWebsocket(ws *WebSocket) {
	t.Lock()
	defer t.Unlock()
	for k := range t.Subs {
		if k == ws.Key {
			delete(t.Subs, ws.Key)
		}
	}
	for k := range t.Pubs {
		if k == ws.Key {
			delete(t.Pubs, ws.Key)
		}
	}
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