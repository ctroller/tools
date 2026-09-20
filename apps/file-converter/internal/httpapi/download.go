package httpapi

import (
	"log/slog"
	"mime"
	"net/http"
	"os"

	"trox.dev/file-converter/internal/task"
)

func (a *API) Download(w http.ResponseWriter, r *http.Request) {
	handle := r.PathValue("handle")
	result, ok := a.intake.Lookup(handle)
	if !ok {
		HttpProblem(w, "", "Not Found", http.StatusNotFound, "Job with handle "+handle+" not found.")
		return
	}

	if result.Status != task.StatusDone {
		HttpProblem(w, "", "Conflict", http.StatusConflict, "Job with handle "+handle+" is not done.")
		return
	}

	f, err := os.Open(result.FilePath)
	if err != nil {
		HttpProblemISE(w, "Failed to read file", err)
		return
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			slog.Error("Failed to close file", "err", err)
		}
	}(f)

	stat, err := f.Stat()
	if err != nil {
		HttpProblemISE(w, "Failed to stat file", err)
		return
	}

	w.Header().Set("Content-Type", string(result.Target))
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment",
		map[string]string{"filename": result.BaseName + result.Target.Ext()}))
	http.ServeContent(w, r, result.BaseName+result.Target.Ext(), stat.ModTime(), f)
}
