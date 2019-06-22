package core

import "encoding/json"

func genResponseData(data interface{}, err error) []byte {
	var respData map[string]interface{}
	if err != nil {
		respData = map[string]interface{}{
			"type":    MTResponse,
			"success": false,
			"message": err.Error(),
		}
	} else {
		respData = map[string]interface{}{
			"type":    MTResponse,
			"success": true,
			"message": data,
		}
	}

	jData, err := json.Marshal(respData)
	FatalErr(err)
	return jData
}
