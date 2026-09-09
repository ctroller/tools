package httpapi

import (
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
		HttpProblem(w, "", "Bad Request", http.StatusBadRequest, err.Error())
		return
	}

	file, handle, err := r.FormFile("file")
	if err != nil {
		HttpProblem(w, "", "Bad Request", http.StatusBadRequest, "Missing file parameter")
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
		HttpProblemISE(w, "Failed to detect mimetype", err)
		return
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		HttpProblemISE(w, "Failed to reset file", err)
		return
	}

	source := convert.MediaType(mtype.String())
	targets, found := registry.Formats()[source]

	if !found {
		HttpProblem(w, "", "Unsupported media type", http.StatusUnsupportedMediaType, "The uploaded file "+handle.Filename+" has an unsupported media type ("+mtype.String()+")")
		return
	}

	dst, err := createDest()
	if err != nil {
		HttpProblemISE(w, "Failed to create file", err)
		return
	}
	defer func(dst multipart.File) {
		err := dst.Close()
		if err != nil {
			slog.Error("Failed to close file", "err", err)
		}
	}(dst)

	if _, err := dst.ReadFrom(file); err != nil {
		HttpProblemISE(w, "Failed to copy file", err)
		err := os.Remove(dst.Name())
		if err != nil {
			slog.Error("Failed to remove file", "err", err)
		}
		return
	}

	res := queue.Prepare(dst.Name(), source)

	RenderJSON(w, Result{Handle: res.JobID, DetectedSource: source, Targets: targets})
}

func createDest() (*os.File, error) {
	if err := os.MkdirAll("uploads", 0755); err != nil {
		return nil, err
	}

	return os.CreateTemp("uploads", "uploaded-*")
}
