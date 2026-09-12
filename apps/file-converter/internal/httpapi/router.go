package httpapi

import (
	"net/http"
	"time"

	"trox.dev/file-converter/internal/convert"
	"trox.dev/file-converter/internal/task"
)

type API struct {
	registry *convert.Registry
	intake   *task.JobIntake
}

func NewRouter(r *convert.Registry, i *task.JobIntake) http.Handler {
	api := &API{registry: r, intake: i}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /formats", api.Formats)
	mux.HandleFunc("POST /files", api.Files)
	mux.HandleFunc("GET /files/{handle}", api.FilesHandle)
	mux.HandleFunc("POST /files/{handle}/convert", api.Convert)
	mux.HandleFunc("GET /files/{handle}/events", func(w http.ResponseWriter, r *http.Request) {
		SSEHandler(w, r, 1*time.Second, api.FilesHandleStream)
	})

	return mux
}
