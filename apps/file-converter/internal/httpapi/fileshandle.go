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
	handle := r.PathValue("handle")
	result, ok := a.intake.Lookup(handle)
	if !ok {
		HttpProblem(w, "", "Not Found", http.StatusNotFound, "Job with handle "+handle+" not found.")
		return
	}

	fhr := toFileHandleResult(result)
	RenderJSON(w, fhr)
}

func (a *API) FilesHandleStream(w http.ResponseWriter, r *http.Request) (done bool) {
	handle := r.PathValue("handle")
	result, ok := a.intake.Lookup(handle)
	if !ok {
		writeSSEData(w, asResponse(FileHandleResult{Status: task.StatusFailed, Error: "Job with handle '" + handle + "' not found."}))
		return true
	}

	writeSSEData(w, toFileHandleResult(result))

	return result.Status.Done()
}

func writeSSEData(w http.ResponseWriter, v Response[FileHandleResult]) {
	payload, err := json.Marshal(v)
	if err != nil {
		slog.Error("Failed to encode SSE payload", "err", err)
		return
	}

	AppendSSE(w, string(payload))
}

func toFileHandleResult(result task.JobResult) Response[FileHandleResult] {
	return asResponse(FileHandleResult{Status: result.Status, Error: result.Error(), Handle: result.JobID})
}

func asResponse(result FileHandleResult) Response[FileHandleResult] {
	return Response[FileHandleResult]{Data: result}
}
