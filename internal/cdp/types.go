package cdp

import "encoding/json"

// Command represents a CDP command sent to the browser
type Command struct {
	ID     int                    `json:"id"`
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// Response represents a CDP response from the browser
type Response struct {
	ID     int             `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *ResponseError  `json:"error,omitempty"`
}

// ResponseError represents an error in a CDP response
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Event represents an unsolicited CDP event from the browser
type Event struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}