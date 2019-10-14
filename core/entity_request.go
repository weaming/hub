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
			return "", fmt.Errorf("message data not provided or type is not in %s", ReprStrArr(MTAll...))
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
