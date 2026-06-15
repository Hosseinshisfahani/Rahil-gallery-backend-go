package response

type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func OK(data interface{}) Envelope {
	return Envelope{Success: true, Data: data}
}

func Fail(code, message string) Envelope {
	return Envelope{
		Success: false,
		Error:   &ErrorBody{Code: code, Message: message},
	}
}
