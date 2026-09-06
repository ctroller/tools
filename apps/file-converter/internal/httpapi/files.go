package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/gabriel-vasile/mimetype"
	"trox.dev/file-converter/internal/convert"
)

type Result struct {
	Handle         string              `json:"handle"`
	DetectedSource convert.MediaType   `json:"detectedSource"`
	Targets        []convert.MediaType `json:"targets"`
}

func Files(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(25 << 20)
	if r.MultipartForm != nil {
		defer func(MultipartForm *multipart.Form) {
			err := MultipartForm.RemoveAll()
			if err != nil {
				slog.Error("Failed to remove multipart form", "err", err)
			}
		}(r.MultipartForm)
	}

	if err != nil {
		slog.Error("Failed to parse multipart form", "err", err)
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		slog.Error("Failed to get file from form", "err", err)
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			slog.Error("Failed to close file", "err", err)
		}
	}(file)

	mtype, err := mimetype.DetectReader(file)
	if err != nil {
		slog.Error("Failed to detect mimetype", "err", err)
		http.Error(w, "Failed to detect mimetype", http.StatusInternalServerError)
		return
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		slog.Error("Failed to reset file", "err", err)
		http.Error(w, "Failed to reset file", http.StatusInternalServerError)
		return
	}

	source := convert.MediaType(mtype.String())
	targets, found := registry.Formats()[source]

	if !found {
		slog.Error("No conversion targets found", "source", source)
		http.Error(w, "No conversion targets found", http.StatusUnsupportedMediaType)
		return
	}

	dst, err := createDest()
	if err != nil {
		slog.Error("Failed to create file", "err", err)
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer func(dst multipart.File) {
		err := dst.Close()
		if err != nil {
			slog.Error("Failed to close file", "err", err)
		}
	}(dst)

	if _, err := dst.ReadFrom(file); err != nil {
		slog.Error("Failed to copy file", "err", err)
		http.Error(w, "Failed to copy file", http.StatusInternalServerError)
		err := os.Remove(dst.Name())
		if err != nil {
			slog.Error("Failed to remove file", "err", err)
		}
		return
	}

	res := queue.Prepare(dst.Name())

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(Result{
		Handle:         res.JobID,
		DetectedSource: source,
		Targets:        targets,
	}); err != nil {
		slog.Error("Failed to encode response", "err", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func createDest() (*os.File, error) {
	if err := os.MkdirAll("uploads", 0755); err != nil {
		return nil, err
	}

	return os.CreateTemp("uploads", "uploaded-*")
}
