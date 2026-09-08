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
}

func FilesHandle(w http.ResponseWriter, r *http.Request) {
	if result := lookup(w, r); result != nil {
		RenderJSON(w, result)
	}
}

func FilesHandleStream(w http.ResponseWriter, r *http.Request) (done bool) {
	handle := r.PathValue("handle")
	if handle == "" {
		writeSSEData(w, FileHandleResult{Status: task.StatusFailed, Error: "Missing handle param."})
		return true
	}

	result, ok := queue.Lookup(handle)
	if !ok {
		writeSSEData(w, FileHandleResult{Status: task.StatusFailed, Error: "Job with handle '" + handle + "' not found."})
		return true
	}

	writeSSEData(w, toFileHandleResult(result))

	return result.Status == task.StatusDone || result.Status == task.StatusFailed
}

func writeSSEData(w http.ResponseWriter, v any) {
	payload, err := json.Marshal(v)
	if err != nil {
		slog.Error("Failed to encode SSE payload", "err", err)
		return
	}

	if _, err := w.Write(append(append([]byte("data: "), payload...), '\n', '\n')); err != nil {
		slog.Error("Failed to write response", "err", err)
	}
}

func lookup(w http.ResponseWriter, r *http.Request) *FileHandleResult {
	handle := r.PathValue("handle")
	if handle == "" {
		HttpProblem(w, "", "Bad Request", http.StatusBadRequest, "Missing handle param.")
		return nil
	}

	result, ok := queue.Lookup(handle)
	if !ok {
		HttpProblem(w, "", "Not Found", http.StatusNotFound, "Job with handle "+handle+" not found.")
		return nil
	}

	fhr := toFileHandleResult(result)
	return &fhr
}

func toFileHandleResult(result task.JobResult) FileHandleResult {
	var err string
	if result.Error != nil {
		err = result.Error.Error()
	}

	return FileHandleResult{Status: task.JobStatus(string(result.Status)), Error: err}
}
