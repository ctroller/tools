package httpapi

import (
	"net/http"
	"time"

	"trox.dev/file-converter/internal/convert"
	"trox.dev/file-converter/internal/task"
)

type API struct {
	registry  *convert.Registry
	queue     *task.Queue
	fileStore *task.FileStore
}

func NewRouter(r *convert.Registry, q *task.Queue, fs *task.FileStore) http.Handler {
	api := &API{registry: r, queue: q, fileStore: fs}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /formats", api.Formats)
	mux.HandleFunc("POST /files", api.Files)
	mux.HandleFunc("GET /files/{handle}", api.FilesHandle)
	mux.HandleFunc("PUT /files/{handle}/convert", api.Convert)
	mux.HandleFunc("GET /files/{handle}/events", func(w http.ResponseWriter, r *http.Request) {
		SSEHandler(w, r, 1*time.Second, api.FilesHandleStream)
	})

	return mux
}
