package httpapi

import (
	"net/http"
)

func (a *API) Formats(w http.ResponseWriter, _ *http.Request) {
	RenderJSON(w, a.registry.Formats())
}
