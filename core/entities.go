package core

import (
	"time"
)

type MessageType string

const (
	MessageTypePlain MessageType = "PLAIN"
	MessageTypeHTML  MessageType = "HTML"
	MessageTypeImage MessageType = "IMAGE"
)

type Message struct {
	Type MessageType
}

type Topic struct {
	Topic     string
	WSSub     map[string]interface{}
	WSPub     map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
	Close     chan (bool)
}

type Hub struct {
	Topics map[string]*Topic
}

func (p *Hub) Append(topic string) {

}
