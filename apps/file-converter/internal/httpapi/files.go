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

func (a *API) Files(w http.ResponseWriter, r *http.Request) {
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
	targets, found := a.registry.Supports(source)
	if !found {
		HttpProblem(w, "", "Unsupported media type", http.StatusUnsupportedMediaType, "The uploaded file "+handle.Filename+" has an unsupported media type ("+mtype.String()+")")
		return
	}

	res, err := a.intake.Submit(file, source)
	if err != nil {
		HttpProblemISE(w, "Failed to submit file", err)
		return
	}

	RenderJSON(w, Response[Result]{Data: Result{Handle: res.JobID, DetectedSource: source, Targets: targets}})
}
