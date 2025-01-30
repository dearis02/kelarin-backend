package types

import (
	"encoding/json"
	"net/http"
)

type ApiResponse struct {
	json.Marshaler
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Errors     []any       `json:"errors,omitempty"`
}

func (r *ApiResponse) MarshalJSON() ([]byte, error) {
	if r.Message == "" {
		r.Message = http.StatusText(r.StatusCode)
	}

	return json.Marshal(r)
}
