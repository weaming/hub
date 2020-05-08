package core

func composeReponse(data interface{}, err error) PushMessageResponse {
	if err != nil {
		return PushMessageResponse{MTResponse, false, err.Error()}
	}
	return PushMessageResponse{MTResponse, true, data}
}

func genResponseData(data interface{}, err error) []byte {
	return ToJSON(composeReponse(data, err))
}
