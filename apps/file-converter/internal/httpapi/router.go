package httpapi

import (
	"net/http"

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

	return mux
}
