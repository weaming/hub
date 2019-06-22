package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

var HUB = &Hub{
	Topics: map[string]*Topic{},
}

type MessageType string

// representation of internal messages
const (
	MessageTypePlain MessageType = "PLAIN"
	MessageTypeHTML  MessageType = "HTML"
	MessageTypeImage MessageType = "IMAGE"
)

// representation of websocket messages
const (
	MTFeedback MessageType = "FEEDBACK"
	MTResponse MessageType = "RESPONSE"
)

type Message struct {
	Type      MessageType `json:"type"`
	Data      string      `json:"data"` // string or base64 of bytes
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
		p.UpdatedAt = time.Now()
	}
}

func (p *Topic) Pub(msg *Message) {
	p.Lock()
	defer p.Unlock()

	if msg.SourceWS != nil {
		ws := msg.SourceWS
		if _, ok := p.Pubs[ws.key]; !ok {
			p.Pubs[ws.key] = ws
			p.UpdatedAt = time.Now()
		}
	}

	c := 0
	for _, sub := range p.Subs {
		// do not send back to self
		if sub != msg.SourceWS {
			go sub.Send(msg)
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

// http client message

type ReqResMessage struct {
	Type MessageType `json:"type"`
	Data string      `json:"data"`
}
type ClientMessage struct {
	Action  string        `json:"action"`
	Topics  []string      `json:"topics"`
	Message ReqResMessage `json:"message"`
}

const (
	ACTION_PUB = "PUB"
	ACTION_SUB = "SUB"
)

func UnmarshalClientMessage(msg []byte) (*ClientMessage, error) {
	clientMsg := &ClientMessage{}
	err := json.Unmarshal(msg, clientMsg)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return clientMsg, nil
}

func (p *ClientMessage) Process(ws *WebSocket) (m string, err error) {
	// p.Message maybe not nil but dereference fail
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		err = fmt.Errorf("%v", r)
	// 	}
	// }()

	topics := p.Topics
	topicsStr := StrArr2Str(topics)

	if len(topics) == 0 {
		return "", errors.New("missing topics")
	}

	switch p.Action {
	case ACTION_PUB:
		message := p.Message
		log.Printf("%+v", message)
		var msg *Message
		if ws != nil {
			msg = &Message{
				Type:      message.Type,
				Data:      message.Data,
				SourceReq: ws.req,
				SourceWS:  ws,
			}
		} else {
			msg = &Message{
				Type: message.Type,
				Data: message.Data,
			}
		}
		for _, topic := range topics {
			ws.Pub(topic, msg)
		}
		return fmt.Sprintf("published on topic %s", topicsStr), nil
	case ACTION_SUB:
		log.Printf("subscribed topics %s", topicsStr)
		if ws == nil {
			return "", fmt.Errorf("HTTP does not support action %s", ACTION_SUB)
		}
		for _, topic := range topics {
			ws.Sub(topic)
		}
		return fmt.Sprintf("subscribed topic %s", topicsStr), nil
	default:
		return "", fmt.Errorf("unsupported action %s", p.Action)
	}
}
