package base

type Response struct {
	Data  interface{}    `json:"data,omitempty"`
	Error *ResponseError `json:"error,omitempty"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

const (
	ResponseErrorCode = iota
	ResponseErrorCodeInternalError
)
