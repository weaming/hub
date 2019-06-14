package core

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

const (
	GET  = "GET"
	POST = "POST"
)

func HTTPPubHandler(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	if req.Method == POST {
		// success
		defer req.Body.Close()
		body, _ := ioutil.ReadAll(req.Body)

		// push into TelegramNotificationBox
		println(body)
		return "success", nil
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	return nil, errors.New("method not allowed")
}

func WSPubHandler(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	return nil, errors.New("method not allowed")
}

func WSSubHandler(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	return nil, errors.New("method not allowed")
}

func APIInterface(fn func(http.ResponseWriter, *http.Request) (interface{}, error), needTopic bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var respData map[string]interface{}
		if needTopic && GetQuery(req, "topic") == "" {
			respData = map[string]interface{}{
				"success": false,
				"data":    "missing query topic",
			}
		} else {
			data, err := fn(w, req)
			if err != nil {
				respData = map[string]interface{}{
					"success": false,
					"data":    err.Error(),
				}
			} else {
				respData = map[string]interface{}{
					"success": true,
					"data":    data,
				}
			}
		}

		jData, err := json.Marshal(respData)
		FatalErr(err)
		w.Write(jData)
	}
}

func ServeHub(listen string) {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("service is healthy"))
	})

	http.HandleFunc("/http/new", APIInterface(HTTPPubHandler, true))
	http.HandleFunc("/ws/pub", APIInterface(WSPubHandler, false))
	http.HandleFunc("/ws/sub", APIInterface(WSSubHandler, false))

	err := http.ListenAndServe(listen, nil)
	FatalErr(err)
}
