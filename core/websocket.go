package core

import (
	"encoding/json"
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
	*sync.Mutex
	key     string
	conn    *websocket.Conn
	req     *http.Request
	Topics  []string
	ErrChan chan error
}

func NewWebsocket(w http.ResponseWriter, r *http.Request) *WebSocket {
	conn, err := upgrader.Upgrade(w, r, nil)
	rv := &WebSocket{
		key:     fmt.Sprintf("%v", conn),
		conn:    conn,
		req:     r,
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
	p.Lock()
	defer p.Unlock()
	// TODO
}

func (p *WebSocket) ProcessMessage() {
	for {
		messageType, msg, err := p.conn.ReadMessage()
		if err != nil {
			p.ErrChan <- err
			return
		}

		var data map[string]interface{}
		switch messageType {
		case websocket.TextMessage:
			// SUB
			// TODO
			data = map[string]interface{}{
				"ok":  true,
				"msg": fmt.Sprintf("subscribed topic %s", string(msg)),
			}

			// PUB
			// TODO
			data = map[string]interface{}{
				"ok":  true,
				"msg": fmt.Sprintf("published on topic %s", string(msg)),
			}
		case websocket.BinaryMessage:
			data = map[string]interface{}{
				"ok":  false,
				"msg": "binary message is not supported",
			}
		}

		// send back
		jData, err := json.Marshal(data)
		FatalErr(err)
		// conn maybe have been closed by manual or client
		if err := p.conn.WriteMessage(websocket.TextMessage, jData); err != nil {
			p.ErrChan <- err
			return
		}
	}
}
