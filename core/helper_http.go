package core

import (
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
