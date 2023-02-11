package main

import (
	"encoding/json"
	"net/http"
)

type JSONResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data",omitempty`
}

// variadic parameter ... -> 0 or any
func (app *application) writeJSON(
	w http.ResponseWriter, status int, data interface{}, headers ...http.Header,
) error {
	out, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	if len(headers) > 0 {
		for _, header := range headers {
			for key, value := range header {
				w.Header()[key] = value
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(out); err != nil {
		return err
	}

	return nil
}
