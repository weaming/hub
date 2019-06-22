package core

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	GET  = "GET"
	POST = "POST"
)

func WSHandler(w http.ResponseWriter, req *http.Request) {
	ws := NewWebsocket(w, req)
	go ws.ProcessMessage()
}

func HTTPHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var data interface{}
	var err error

	if req.Method == POST {
		defer req.Body.Close()
		body, _ := ioutil.ReadAll(req.Body)
		clientMsg, e := UnmarshalClientMessage(body)
		if e != nil {
			err = e
		} else {
			data, err = clientMsg.Process(nil)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		err = errors.New("method not allowed")
	}

	w.Write(genResponseData(data, err))
}

func ServeHub(listen string) {
	log.Printf("serve http on %s", listen)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("service is healthy"))
	})

	http.HandleFunc("/http", HTTPHandler)
	http.HandleFunc("/ws", WSHandler)

	err := http.ListenAndServe(listen, nil)
	FatalErr(err)
}
