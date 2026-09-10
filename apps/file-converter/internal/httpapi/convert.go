package httpapi

import (
	"errors"
	"net/http"

	"trox.dev/file-converter/internal/common"
	"trox.dev/file-converter/internal/convert"
	"trox.dev/file-converter/internal/task"
)

func (a *API) Convert(w http.ResponseWriter, r *http.Request) {
	handle := r.PathValue("handle")
	if handle == "" {
		HttpProblem(w, "", "Bad Request", http.StatusBadRequest, "Missing handle param.")
		return
	}

	target := convert.MediaType(r.URL.Query().Get("target"))
	if target == "" {
		HttpProblem(w, "", "Bad Request", http.StatusBadRequest, "Missing target param.")
		return
	}

	if err := a.intake.StartJob(handle, target); err != nil {
		var notFoundErr common.NotFoundErr
		var illegalStateErr common.IllegalStateErr
		var queueFullErr task.QueueFullErr
		switch {
		case errors.As(err, &notFoundErr):
			HttpProblem(w, "", "Not Found", http.StatusNotFound, err.Error())
		case errors.As(err, &illegalStateErr):
			HttpProblem(w, "", "Conflict", http.StatusConflict, err.Error())
		case errors.As(err, &queueFullErr):
			HttpProblem(w, "", "Unable to process, try again", http.StatusServiceUnavailable, err.Error())
		default:
			HttpProblemISE(w, "Unable to start job", err)
		}
		return
	}

	w.Header().Set("Location", "/files/"+handle)
	w.WriteHeader(http.StatusAccepted)
}
