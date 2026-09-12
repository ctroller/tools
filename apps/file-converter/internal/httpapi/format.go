package httpapi

import (
	"net/http"

	"trox.dev/file-converter/internal/convert"
)

func (a *API) Formats(w http.ResponseWriter, _ *http.Request) {
	RenderJSON(w, Response[map[convert.MediaType][]convert.MediaType]{Data: a.registry.Formats()})
}
