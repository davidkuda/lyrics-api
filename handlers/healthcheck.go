package handlers

import (
	"net/http"
)

func (app *Application) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	js := `{"status": "available", "environment": "development", "version": "1.0.0"}`

	w.Header().Set("Content-Type", "application/json")

	w.Write([]byte(js))
}
