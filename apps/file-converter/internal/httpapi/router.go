package httpapi

import (
	"net/http"

	"trox.dev/file-converter/internal/convert"
)

func NewRouter(registry *convert.Registry) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/formats", func(w http.ResponseWriter, r *http.Request) {
		Formats(w, registry)
	})

	return mux
}
