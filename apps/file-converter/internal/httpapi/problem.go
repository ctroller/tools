package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

type ProblemDetails struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
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

	problem := ProblemDetails{
		Type:   t,
		Title:  title,
		Status: status,
		Detail: detail,
	}

	Render(w, "application/problem+json", status, func(w io.Writer) error {
		return json.NewEncoder(w).Encode(problem)
	})
}
