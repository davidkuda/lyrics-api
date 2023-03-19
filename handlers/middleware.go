package handlers

import (
	"net/http"
)

func (app *Application) EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		origin := r.Header.Get("Origin")

		if origin != "" {
			for i := range app.CORS.TrustedOrigins {
				if origin == app.CORS.TrustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, credentials")
					break
				}
			}
		}

		// Browsers will send a CORS preflight request before
		// the actual request with the OPTIONS method and will
		// only continue with the next request if the preflight
		// returns a success.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
