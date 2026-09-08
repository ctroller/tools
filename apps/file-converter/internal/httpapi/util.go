package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

func Render(w http.ResponseWriter, cType string, status int, handler func(w io.Writer) error) {
	var buf bytes.Buffer
	if err := handler(&buf); err != nil {
		slog.Error("Unable to render response", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", cType)
	w.WriteHeader(status)
	if _, err := w.Write(buf.Bytes()); err != nil {
		slog.Error("Unable to write response", "err", err)
	}
}

func RenderJSON(w http.ResponseWriter, data any) {
	Render(w, "application/json", http.StatusOK, func(w io.Writer) error {
		return json.NewEncoder(w).Encode(data)
	})
}

func RenderJSONStatus(w http.ResponseWriter, data any, cType string, status int) {
	Render(w, cType, status, func(w io.Writer) error {
		return json.NewEncoder(w).Encode(data)
	})
}
