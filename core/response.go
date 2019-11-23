package core

func composeReponse(data interface{}, err error) map[string]interface{} {
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
	return respData
}
func genResponseData(data interface{}, err error) []byte {
	return ToJSON(composeReponse(data, err))
}
