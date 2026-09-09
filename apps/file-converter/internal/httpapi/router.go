package httpapi

import (
	"net/http"
	"time"

	"trox.dev/file-converter/internal/convert"
	"trox.dev/file-converter/internal/task"
)

var registry *convert.Registry
var queue *task.Queue

func NewRouter(r *convert.Registry, q *task.Queue) http.Handler {
	registry = r
	queue = q
	mux := http.NewServeMux()

	mux.HandleFunc("GET /formats", Formats)
	mux.HandleFunc("POST /files", Files)
	mux.HandleFunc("GET /files/{handle}", FilesHandle)
	mux.HandleFunc("PUT /files/{handle}/convert", Convert)
	mux.HandleFunc("GET /files/{handle}/events", func(w http.ResponseWriter, r *http.Request) {
		SSEHandler(w, r, 1*time.Second, FilesHandleStream)
	})

	return mux
}
