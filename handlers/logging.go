package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

type requestLog struct {
	URL      string `json:"url"`
	Method   string `json:"method"`
	BodySize int64  `json:"content_length"`
	Protocol string `json:"protocol"`
}

func (a *Application) LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logRequest(r, a.Config.Logger)
		next.ServeHTTP(w, r)
	})
}

func logRequest(r *http.Request, logger *log.Logger) {
	l := requestLog{
		URL:      r.URL.String(),
		Method:   r.Method,
		BodySize: r.ContentLength,
		Protocol: r.Proto,
	}

	j, err := json.Marshal(&l)
	if err != nil {
		panic(err)
	}

	logger.Println(string(j))
}

