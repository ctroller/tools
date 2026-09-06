package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func Formats(w http.ResponseWriter) {
	formats := registry.Formats()
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(formats); err != nil {
		slog.Error("Failed to encode formats", "err", err)
		http.Error(w, "Failed to encode formats", http.StatusInternalServerError)
		return
	}
}
