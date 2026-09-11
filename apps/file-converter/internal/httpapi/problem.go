package httpapi

import (
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

	RenderJSONStatus(w, problem, "application/problem+json", status)
}
