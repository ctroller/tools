package httpapi

import (
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"

	"github.com/gabriel-vasile/mimetype"
	"trox.dev/file-converter/internal/convert"
)

type Result struct {
	Handle         string              `json:"handle"`
	DetectedSource convert.MediaType   `json:"detectedSource"`
	Targets        []convert.MediaType `json:"targets"`
}

// detectAndSubmit detects the media type of file, checks it is supported, and submits it as a
// new conversion job. unsupportedMediaType builds the error detail for an unsupported type.
func (a *API) detectAndSubmit(w http.ResponseWriter, file io.ReadSeeker, unsupportedMediaType func(t string) string) (Result, bool) {
	t, err := mimetype.DetectReader(file)
	if err != nil {
		HttpProblemISE(w, "Failed to detect mimetype", err)
		return Result{}, false
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		HttpProblemISE(w, "Failed to reset file", err)
		return Result{}, false
	}

	source := convert.MediaType(t.String())
	targets, found := a.registry.Supports(source)
	if !found {
		HttpProblem(w, "", "Unsupported media type", http.StatusUnsupportedMediaType, unsupportedMediaType(t.String()))
		return Result{}, false
	}

	res, err := a.intake.Submit(file, source)
	if err != nil {
		HttpProblemISE(w, "Failed to submit file", err)
		return Result{}, false
	}

	return Result{Handle: res.JobID, DetectedSource: source, Targets: targets}, true
}

func (a *API) Files(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, a.config.MaxFileSizeBytes)
	err := r.ParseMultipartForm(a.config.MaxFileSizeBytes)
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

	result, ok := a.detectAndSubmit(w, file, func(t string) string {
		return "The uploaded file " + handle.Filename + " has an unsupported media type (" + t + ")"
	})
	if !ok {
		return
	}

	RenderJSON(w, Response[Result]{Data: result})
}
