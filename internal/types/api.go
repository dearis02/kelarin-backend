package types

type ApiResponse struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Errors     []any       `json:"errors,omitempty"`
}
