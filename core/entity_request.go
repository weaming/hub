package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

type ReqResMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func (p *ReqResMessage) Str() string {
	if InStrArr(p.Type, MTAll...) {
		return p.Data
	}
	return ""
}

type PubRequest struct {
	Action  string        `json:"action"`
	Topics  []string      `json:"topics"`
	Subs    []string      `json:"subs"`
	Message ReqResMessage `json:"message"`
	Hub     *Hub          `json:"-"`
}

const (
	ActionPub = "PUB"
	ActionSub = "SUB"
)

func UnmarshalClientMessage(msg []byte, hub *Hub) (*PubRequest, error) {
	clientMsg := &PubRequest{Hub: hub}
	err := json.Unmarshal(msg, clientMsg)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return clientMsg, nil
}

func (p *PubRequest) Pub(topic string, msg *PubMessage) {
	p.Hub.Pub(topic, msg)
}

func (p *PubRequest) Process(ws *WebSocket) (m string, err error) {
	// optional ws, nil stands for HTTP client PUBlished a message

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
			return "", fmt.Errorf("message data not provided or type is not in %s", ReprStrArr(MTAll...))
		}

		var msg *PubMessage
		if ws != nil {
			msg = &PubMessage{
				Type:      message.Type,
				Data:      message.Data,
				SourceReq: ws.req,
				SourceWS:  ws,
			}
			for _, topic := range topics {
				// publish through the ws
				ws.Pub(topic, msg)
			}
		} else {
			msg = &PubMessage{
				Type: message.Type,
				Data: message.Data,
			}
			for _, topic := range topics {
				// message can publish to HUB directly
				p.Pub(topic, msg)
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
