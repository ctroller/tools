package httpapi

import (
	"log/slog"
	"net/http"
	"time"
)

func SSEHandler(w http.ResponseWriter, r *http.Request, d time.Duration, handle func(w http.ResponseWriter, r *http.Request) (done bool)) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	gone := r.Context().Done()
	rc := http.NewResponseController(w)
	t := time.NewTicker(d)
	defer t.Stop()
	for {
		select {
		case <-gone:
			return
		case <-t.C:
			done := handle(w, r)
			if err := rc.Flush(); err != nil {
				slog.Error("Failed to flush response", "err", err)
				return
			}
			if done {
				return
			}
		}
	}
}
