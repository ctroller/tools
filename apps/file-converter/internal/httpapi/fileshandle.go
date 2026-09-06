package httpapi

import (
	"net/http"
)

type FileHandleResult struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func FilesHandle(w http.ResponseWriter, r *http.Request) {
	handle := r.PathValue("handle")
	if handle == "" {
		HttpProblem(w, "", "Bad Request", http.StatusBadRequest, "Missing handle param.")
		return
	}

	result, ok := queue.Lookup(handle)
	if !ok {
		HttpProblem(w, "", "Not Found", http.StatusNotFound, "Job with handle "+handle+" not found.")
		return
	}

	var err string
	if result.Error != nil {
		err = result.Error.Error()
	}

	RenderJSON(w, FileHandleResult{Status: string(result.Status), Error: err})
}
