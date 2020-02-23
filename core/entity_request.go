package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

type PubRequest struct {
	Action  string     `json:"action"`
	Topics  []string   `json:"topics"`
	Message PubMessage `json:"message"`
	hub     *Hub
}

const (
	ActionPub = "PUB"
	ActionSub = "SUB"
)

func UnmarshalClientMessage(msg []byte, hub *Hub) (*PubRequest, error) {
	clientMsg := &PubRequest{hub: hub}
	err := json.Unmarshal(msg, clientMsg)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return clientMsg, nil
}

func (p *PubRequest) Pub(topic string, msg *PubMessage) {
	p.hub.Pub(topic, msg)
}

func (p *PubRequest) Process(ws *WebSocket) (m string, err error) {
	// optional ws, nil stands for a message published by HTTP client
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
			return "", fmt.Errorf("message data not provided or type is not in %s", ReprStrArr(MTAll...))
		}

		var msg *PubMessage
		if ws != nil {
			message.SourceReq = ws.req
			message.SourceWS = ws
			for _, topic := range topics {
				// publish through the ws
				ws.Pub(topic, msg)
			}
		} else {
			for _, topic := range topics {
				// message can publish to HUB directly
				p.Pub(topic, &message)
			}
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
