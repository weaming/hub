package core

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WebSocket struct {
	sync.Mutex
	conn    *websocket.Conn
	req     *http.Request
	Key     string     `json:"key"`
	Topics  []string   `json:"topics"`
	ErrChan chan error `json:"-"`
}

func NewWebsocket(w http.ResponseWriter, r *http.Request) *WebSocket {
	conn, err := upgrader.Upgrade(w, r, nil)
	rv := &WebSocket{
		Key:     Sha256([]byte(fmt.Sprintf("%+v", conn))),
		conn:    conn,
		req:     r,
		Topics:  []string{},
		ErrChan: make(chan error, 1),
	}
	if err != nil {
		rv.ErrChan <- err
	}
	go rv.ProcessError()
	go rv.ProcessMessage()
	return rv
}

func (p *WebSocket) ProcessError() {
	for {
		err := <-p.ErrChan
		if err != nil {
			log.Printf("[WebSocet] %v", err)
			p.Close()
			return
		}
	}
}

func (p *WebSocket) Close() {
	p.conn.Close()
}

func (p *WebSocket) Sub(topic string) {
	if !InStrArr(topic, p.Topics...) {
		p.Topics = append(p.Topics, topic)
		HUB.Sub(topic, p)
	}
}

func (p *WebSocket) Pub(topic string, msg *Message) {
	HUB.Pub(topic, msg)
}

func (p *WebSocket) Send(msg *Message) {
	bytes := ToJSON(ReqResMessage{
		Type: msg.Type,
		Data: msg.Data,
	})
	err := p.WriteSafe(bytes)
	if err != nil {
		p.ErrChan <- err
	}
}

func (p *WebSocket) WriteSafe(bytes []byte) error {
	p.Lock()
	defer p.Unlock()
	return p.conn.WriteMessage(websocket.TextMessage, bytes)
}

func (p *WebSocket) ProcessMessage() {
	for {
		messageType, msg, e := p.conn.ReadMessage()
		if e != nil {
			p.ErrChan <- e
			return
		}

		var data interface{}
		var err error

		switch messageType {
		case websocket.TextMessage:
			clientMsg, e := UnmarshalClientMessage(msg)
			if e != nil {
				err = e
			} else {
				data, err = clientMsg.Process(p)
			}
		case websocket.BinaryMessage:
			err = fmt.Errorf("binary message is not supported")
		}

		jData := genResponseData(data, err)
		err = p.WriteSafe(jData)
		if err != nil {
			p.ErrChan <- err
			// conn maybe have been closed by manual or client
			return
		}
	}
}
