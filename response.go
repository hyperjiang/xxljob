package xxljob

import "encoding/json"

// Response is the response format.
type Response struct {
	Code    int         `json:"code"` // 200 means success, other means failed
	Msg     string      `json:"msg"`  // error message
	Content interface{} `json:"content,omitempty"`
}

// String marshals the response into a json string.
func (res Response) String() string {
	b, _ := json.Marshal(res)
	return string(b)
}

// NewSuccResponse returns a default success response.
func NewSuccResponse() *Response {
	return &Response{Code: successCode}
}

// NewErrorResponse returns an error response.
func NewErrorResponse(msg string) *Response {
	return &Response{Code: failureCode, Msg: msg}
}

// LogResult is the response format for log content.
type LogResult struct {
	FromLineNum int    `json:"fromLineNum"`
	ToLineNum   int    `json:"toLineNum"`
	LogContent  string `json:"logContent"`
	IsEnd       bool   `json:"isEnd"`
}
