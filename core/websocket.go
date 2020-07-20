package core

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WebSocket struct {
	sync.Mutex
	conn      *websocket.Conn
	req       *http.Request
	ID        string     `json:"id"`
	Topics    []string   `json:"topics"` // subscribed topics
	ErrChan   chan error `json:"-"`
	CreatedAt time.Time  `json:"created_at"`
	Hub       *Hub       `json:"-"`
}

func NewWebsocket(c *gin.Context) *WebSocket {
	w := c.Writer
	r := c.Request
	conn, err := upgrader.Upgrade(w, r, nil)
	rv := &WebSocket{
		ID:        Sha256([]byte(fmt.Sprintf("%+v", conn))),
		conn:      conn,
		req:       r,
		Topics:    []string{},
		ErrChan:   make(chan error, 1),
		CreatedAt: time.Now(),
		Hub:       getHub(c),
	}
	if err != nil {
		rv.ErrChan <- err
	}
	// https://godoc.org/github.com/gorilla/websocket#hdr-Concurrency
	go rv.ProcessError()
	go rv.ProcessMessage()
	return rv
}

func (w *WebSocket) ProcessError() {
	for {
		err := <-w.ErrChan
		if err != nil {
			log.Printf("[WebSocket] %v", err)
			w.Close()
			return
		}
	}
}

func (w *WebSocket) Close() {
	w.conn.Close()
	for _, t := range w.Hub.Topics {
		t.dereferenceWebsocket(w)
	}
}

func (w *WebSocket) Sub(topic string) {
	if !InStrArr(topic, w.Topics...) {
		w.Topics = append(w.Topics, topic)
		w.Hub.Sub(topic, w)
		w.WriteSafe(ToJSON(PushMessageFeedback{
			Type:    MTFeedback,
			Message: fmt.Sprintf(`subscribed on topic "%s"`, topic),
		}))
	}
}

func (w *WebSocket) Pub(topic string, msg *PubMessage) {
	w.Hub.Pub(topic, msg)
}

//  send message to subscribers
func (w *WebSocket) send(topic string, msg *PubMessage) {
	bytes := ToJSON(PushMessage{
		Type:    MTMessage,
		Topic:   topic,
		Message: msg,
	})
	err := w.WriteSafe(bytes)
	if err != nil {
		w.ErrChan <- err
	}
}

func (w *WebSocket) WriteSafe(bytes []byte) error {
	w.Lock()
	defer w.Unlock()
	return w.conn.WriteMessage(websocket.TextMessage, bytes)
}

// call msg.Process() and detect errors
func (w *WebSocket) ProcessMessage() {
	for {
		messageType, msg, e := w.conn.ReadMessage()
		if e != nil {
			w.ErrChan <- e
			return
		}

		var data interface{}
		var err error

		switch messageType {
		case websocket.TextMessage:
			if clientMsg, err := UnmarshalClientMessage(msg, w.Hub); err == nil {
				data, err = clientMsg.Process(w)
			}
		case websocket.BinaryMessage:
			err = fmt.Errorf("binary message is not supported")
		}

		if err != nil {
			log.Println("parse ws msg fail:", err)
			continue
		}

		if err = w.WriteSafe(genResponseData(data, err)); err != nil {
			w.ErrChan <- err
			// conn maybe have been closed by manual or client
			return
		}
	}
}
