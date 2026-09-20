package httpapi

import (
	"log/slog"
	"net/http"
)

func (a *API) DeleteHandle(w http.ResponseWriter, r *http.Request) {
	handle := r.PathValue("handle")
	if handle == "" {
		HttpProblem(w, "", "Bad Request", http.StatusBadRequest, "Missing handle param.")
		return
	}

	result, found, deleted := a.intake.Delete(handle)
	if !found {
		HttpProblem(w, "", "Not Found", http.StatusNotFound, "Job not found.")
		return
	}
	if !deleted {
		HttpProblem(w, "", "Conflict", http.StatusConflict, "Job is not done.")
		return
	}

	slog.Info("Deleting job", "handle", handle, "status", result.Status)
	if result.FilePath != "" {
		a.fs.Delete(result.FilePath)
	}

	w.WriteHeader(http.StatusAccepted)
}
