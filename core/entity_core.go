package core

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Hub for public
var HUBPublic = NewHub()

// Hub for share
var HUBShare = NewHub()

const GlobalTopicID = "global"

// types of internal messages
const (
	MTPlain    string = "PLAIN"
	MTMarkdown string = "MARKDOWN"
	MTJSON     string = "JSON"
	MTHTML     string = "HTML"
	MTPhoto    string = "PHOTO"
	MTVideo    string = "VIDEO"
)

var MTAll = []string{MTPlain, MTMarkdown, MTJSON, MTHTML, MTPhoto, MTVideo}

// types of websocket messages
const (
	MTFeedback string = "FEEDBACK" // used for async event feedback
	MTResponse string = "RESPONSE" // used for message response
	MTMessage  string = "MESSAGE"  // used for publish messages
)

type RawMessage struct {
	Type    string `json:"type"`    // required
	Data    string `json:"data"`    // required, string or base64 of bytes
	Caption string `json:"caption"` // optional
}

func (p *RawMessage) isMedia() bool {
	return p.Type == MTPhoto || p.Type == MTVideo
}

// http client message
type PubMessage struct {
	Type         string        `json:"type"`
	Data         string        `json:"data"` // string or base64 of bytes
	Caption      string        `json:"caption"`
	ExtendedData []RawMessage  `json:"extended_data"` // string or base64 of bytes, for sending multiple photos
	SourceReq    *http.Request `json:"-"`
	SourceWS     *WebSocket    `json:"-"`
}

func (p *PubMessage) Str() string {
	if InStrArr(p.Type, MTAll...) {
		return p.Data
	}
	return ""
}
func (p *PubMessage) isMedia() bool {
	return p.Type == MTPhoto || p.Type == MTVideo
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
	if _, ok := t.Subs[ws.ID]; !ok {
		t.Subs[ws.ID] = ws
		t.UpdatedAt = time.Now()
	}
}

func (t *Topic) Pub(msg *PubMessage) {
	t.Lock()
	defer t.Unlock()

	if msg.SourceWS != nil {
		ws := msg.SourceWS
		if _, ok := t.Pubs[ws.ID]; !ok {
			t.Pubs[ws.ID] = ws
			t.UpdatedAt = time.Now()
		}
	}

	// save into in-memory buffers
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
		if k == ws.ID {
			delete(t.Subs, ws.ID)
		}
	}
	for k := range t.Pubs {
		if k == ws.ID {
			delete(t.Pubs, ws.ID)
		}
	}
}

type Hub struct {
	sync.Mutex
	Topics map[string]*Topic `json:"topics"`
}

func NewHub() *Hub {
	return &Hub{
		Topics: map[string]*Topic{},
	}
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

func (p *Hub) Pub(topic string, msg *PubMessage) {
	tpc := p.GetTopic(topic)
	tpc.Pub(msg)
}
