package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"trox.dev/file-converter/internal/task"
)

type FileHandleResult struct {
	Status task.JobStatus `json:"status"`
	Error  string         `json:"error,omitempty"`
	Handle string         `json:"handle"`
}

func (a *API) FilesHandle(w http.ResponseWriter, r *http.Request) {
	if result := a.lookupHandle(w, r); result != nil {
		RenderJSON(w, result)
	}
}

func (a *API) FilesHandleStream(w http.ResponseWriter, r *http.Request) (done bool) {
	handle := r.PathValue("handle")
	if handle == "" {
		writeSSEData(w, FileHandleResult{Status: task.StatusFailed, Error: "Missing handle param."})
		return true
	}

	result, ok := a.intake.Lookup(handle)
	if !ok {
		writeSSEData(w, FileHandleResult{Status: task.StatusFailed, Error: "Job with handle '" + handle + "' not found."})
		return true
	}

	writeSSEData(w, toFileHandleResult(result))

	return result.Status.Done()
}

func writeSSEData(w http.ResponseWriter, v any) {
	payload, err := json.Marshal(v)
	if err != nil {
		slog.Error("Failed to encode SSE payload", "err", err)
		return
	}

	AppendSSE(w, string(payload))
}
