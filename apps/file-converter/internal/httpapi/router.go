package httpapi

import (
	"net/http"
	"time"

	"trox.dev/file-converter/internal/config"
	"trox.dev/file-converter/internal/convert"
	"trox.dev/file-converter/internal/task"
)

type API struct {
	registry *convert.Registry
	intake   *task.JobIntake
	fs       *task.FileStore
	config   *config.Config
}

func NewRouter(c *config.Config, r *convert.Registry, i *task.JobIntake, fs *task.FileStore) http.Handler {
	api := &API{config: c, registry: r, intake: i, fs: fs}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /formats", api.Formats)
	mux.HandleFunc("POST /files", api.Files)
	mux.HandleFunc("GET /download-url", api.DownloadUrl)
	mux.HandleFunc("GET /files/{handle}", api.FilesHandle)
	mux.HandleFunc("DELETE /files/{handle}", api.DeleteHandle)
	mux.HandleFunc("POST /files/{handle}/convert", api.Convert)
	mux.HandleFunc("GET /files/{handle}/events", func(w http.ResponseWriter, r *http.Request) {
		SSEHandler(w, r, 1*time.Second, api.FilesHandleStream)
	})

	return mux
}
