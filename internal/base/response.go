package base

// Response represents the response contract
// swagger:response Response
type Response struct {
	Data  any            `json:"data,omitempty"`
	Error *ResponseError `json:"error,omitempty"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
