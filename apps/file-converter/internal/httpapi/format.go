package httpapi

import (
	"net/http"
)

func Formats(w http.ResponseWriter, _ *http.Request) {
	RenderJSON(w, registry.Formats())
}
