package core

import (
	"bytes"
	"encoding/json"
	"net/http"
)

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
