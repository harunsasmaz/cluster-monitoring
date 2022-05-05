package server

import (
	"encoding/json"
	"fmt"
)

type HttpResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Success bool        `json:"success"`
}

type HttpErrorResponse []byte

func HttpError(format string, args ...interface{}) HttpErrorResponse {
	err := fmt.Sprintf(format, args...)

	js, _ := json.Marshal(&HttpResponse{
		Error:   err,
		Success: false,
	})

	return js
}
