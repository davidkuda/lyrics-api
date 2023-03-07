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
	env := envelope{
		"status": "available",
		"system_information": map[string]string{
			"environment": "development",
			"version":     "0.0.1",
		},
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
