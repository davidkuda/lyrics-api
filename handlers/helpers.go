package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type JSONResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data",omitempty`
}

// variadic parameter ... -> 0 or any
func (app *Application) writeJSON(
	w http.ResponseWriter, status int, data interface{}, headers ...http.Header,
) error {
	out, err := json.MarshalIndent(data, "", "  ")
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
	w.Write(out)

	return nil
}

func (app *Application) readJSON(
	w http.ResponseWriter, r *http.Request, data interface{},
) error {
	maxBytes := 1024 * 1024 // one megabyte
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(data); err != nil {
		return err
	}

	// try to decode into a throwaway variable
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("Body mustg only contain a single JSON struct")
	}

	return nil
}

func (app *Application) errorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest
	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload JSONResponse
	payload.Error = true
	payload.Message = err.Error()

	return app.writeJSON(w, statusCode, payload)
}
