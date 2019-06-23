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

const GlobalTopicID = "global"

type MessageType string

// representation of internal messages
const (
	MTPlain MessageType = "PLAIN"
	MTJSON  MessageType = "JSON"
	MTHTML  MessageType = "HTML"
	MTImage MessageType = "IMAGE"
)

var MTAll = []MessageType{MTPlain, MTJSON, MTHTML, MTImage}

// representation of websocket messages
const (
	MTFeedback MessageType = "FEEDBACK" // used for async event feedback
	MTResponse MessageType = "RESPONSE" // used for message response
	MTMessage  MessageType = "MESSAGE"  // used for publish messages
)

type Message struct {
	Type      MessageType `json:"type"`
	Data      string      `json:"data"` // string or base64 of bytes
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

	c := 0
	for _, sub := range p.Subs {
		// do not send back to self
		if sub != msg.SourceWS {
			go sub.Send(msg, p.Topic)
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

// http client message

type ReqResMessage struct {
	Type MessageType `json:"type"`
	Data string      `json:"data"`
}

func (p *ReqResMessage) Str() string {
	if In(p.Type, MTAll...) {
		return p.Data
	}
	return ""
}

type Request struct {
	Action  string        `json:"action"`
	Topics  []string      `json:"topics"`
	Message ReqResMessage `json:"message"`
}

const (
	ActionPub = "PUB"
	ActionSub = "SUB"
)

func UnmarshalClientMessage(msg []byte) (*Request, error) {
	clientMsg := &Request{}
	err := json.Unmarshal(msg, clientMsg)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return clientMsg, nil
}

func (p *Request) Process(ws *WebSocket) (m string, err error) {
	// p.Message maybe not nil but dereference fail
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		err = fmt.Errorf("%v", r)
	// 	}
	// }()

	topics := p.Topics
	topicsStr := ReprStrArr(topics...)

	if len(topics) == 0 {
		return "", errors.New("missing topics")
	}

	switch p.Action {
	case ActionPub:
		message := p.Message
		log.Printf("%+v", message)
		if p.Message.Str() == "" {
			return "", fmt.Errorf("message data not provided or type is not in %s", ReprArr(MTAll...))
		}

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
		return fmt.Sprintf("publish requests on topics %s are processing", topicsStr), nil
	case ActionSub:
		if ws == nil {
			return "", fmt.Errorf("HTTP does not support action %s", ActionSub)
		}
		for _, topic := range topics {
			ws.Sub(topic)
		}
		resText := fmt.Sprintf("subscribe requests on topics %s are processing", topicsStr)
		log.Println(resText)
		return resText, nil
	default:
		return "", fmt.Errorf("unsupported action %s", p.Action)
	}
}
