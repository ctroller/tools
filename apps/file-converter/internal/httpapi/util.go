package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func Render(w http.ResponseWriter, cType string, status int, handler func(w http.ResponseWriter) error) {
	w.Header().Set("Content-Type", cType)
	w.WriteHeader(status)

	if err := handler(w); err != nil {
		slog.Error("Unable to render response", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func RenderJSON(w http.ResponseWriter, data any) {
	Render(w, "application/json", http.StatusOK, func(w http.ResponseWriter) error {
		return json.NewEncoder(w).Encode(data)
	})
}

func RenderJSONStatus(w http.ResponseWriter, data any, cType string, status int) {
	Render(w, cType, status, func(w http.ResponseWriter) error {
		return json.NewEncoder(w).Encode(data)
	})
}
