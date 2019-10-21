package core

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

const (
	GET  = "GET"
	POST = "POST"
)

func WSHandler(w http.ResponseWriter, req *http.Request) {
	ws := NewWebsocket(w, req)
	ws.WriteSafe(genResponseData("connected", nil))
	ws.Sub(GlobalTopicID)
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
		topic := GetQuery(req, "topic")
		amount := GetQuery(req, "amount")
		if topic == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(genResponseData(data, errors.New("missing topic")))
			return
		}
		if amount == "" {
			amount = "10"
		}
		amountN, err := strconv.Atoi(amount)
		if err != nil {
			w.Write(genResponseData(data, err))
			return
		}

		dataBytes := BufGetN(topic, amountN)
		_data := []string{}
		for _, x := range dataBytes {
			s := string(x)
			_data = append(_data, s)
		}

		data = map[string]interface{}{
			"data":  _data,
			"count": len(_data),
		}
		w.Write(genResponseData(data, err))
		return
	}

	w.Write(genResponseData(data, err))
}

func StatusHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(genResponseData(HUB, nil))
}

func ServeHub(listen string) {
	log.Printf("serve http on %s", listen)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Write(ToJSON(map[string]string{
			"status":      "service is healthy",
			"source_code": "https://github.com/weaming/hub",
		}))
	})

	http.HandleFunc("/http", HTTPHandler)
	http.HandleFunc("/ws", WSHandler)
	http.HandleFunc("/status", StatusHandler)

	err := http.ListenAndServe(listen, nil)
	FatalErr(err)
}
