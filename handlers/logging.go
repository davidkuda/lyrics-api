package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/davidkuda/lyricsapi/config"
)

// todo: log via middleware, not inside handlers
// ? how can you write logs to a file? can you write to stdout and to a file? (i.e. to multiple files?)
type requestLog struct {
	URL      string `json:"url"`
	Method   string `json:"method"`
	BodySize int64  `json:"content_length"`
	Protocol string `json:"protocol"`
}

func logRequest(r *http.Request, cfg *config.AppConfig) {
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
	cfg.Logger.Println(string(j))
}

