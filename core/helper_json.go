package core

import "encoding/json"

func ToJSON(v interface{}) []byte {
	bytes, err := json.Marshal(v)
	FatalErr(err)
	return bytes
}
