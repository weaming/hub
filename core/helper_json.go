package core

import (
	"encoding/json"
	"os"

	"github.com/bitly/go-simplejson"
)

func ToJSON(v interface{}) []byte {
	bytes, err := json.Marshal(v)
	FatalErr(err)
	return bytes
}

func ToJSONStr(v interface{}) string {
	return string(ToJSON(v))
}

// https://godoc.org/github.com/bitly/go-simplejson
func LoadJSON(jsonPath string) *simplejson.Json {
	jsonFile, err := os.Open(jsonPath)
	FatalErr(err)
	js, err := simplejson.NewFromReader(jsonFile)
	FatalErr(err)
	return js
}

func MustMapStr(data map[string]interface{}) map[string]string {
	rv := map[string]string{}
	for k, v := range data {
		rv[k] = v.(string)
	}
	return rv
}
