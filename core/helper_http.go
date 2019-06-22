package core

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
)

func GetMessageIP(req *http.Request) string {
	realIP := req.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	return strings.Split(req.RemoteAddr, ":")[0]
}

func GetQuery(req *http.Request, name string) string {
	return req.URL.Query().Get(name)
}

func PostJSON(api string, data map[string]interface{}) (map[string]interface{}, error) {
	bytesRepresentation, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(api, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}
