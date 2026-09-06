package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
	Data     any    `json:"data,omitempty"`
}

func HttpProblemISE(w http.ResponseWriter, errMsg string, err error) {
	HttpProblem(w, "", "Internal Server Error", http.StatusInternalServerError, "")
	slog.Error(errMsg, "err", err)
}

func HttpProblem(w http.ResponseWriter, pType, title string, status int, detail string) {
	t := pType
	if pType == "" {
		t = "about:blank"
	}

	problem := Problem{
		Type:     t,
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: "",
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(problem); err != nil {
		slog.Error("failed to encode problem", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
