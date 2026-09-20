package task

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"uuid"

	"trox.dev/file-converter/internal/common"
	"trox.dev/file-converter/internal/convert"
)

type JobIntake struct {
	store     *StatusStore
	registry  *convert.Registry
	queue     *Queue
	fileStore *FileStore
}

func NewJobIntake(registry *convert.Registry, queue *Queue, store *StatusStore, fileStore *FileStore) *JobIntake {
	return &JobIntake{
		store, registry, queue, fileStore,
	}
}

func (intake *JobIntake) prepare(path, id string, src convert.MediaType, originalFileName string) JobResult {
	originalFileName = strings.TrimSuffix(originalFileName, filepath.Ext(originalFileName))
	if originalFileName == "" || originalFileName == "/" || originalFileName == "." {
		originalFileName = "download"
	}

	result := JobResult{
		JobID:    id,
		Status:   StatusUploaded,
		FilePath: path,
		Source:   src,
		BaseName: originalFileName,
	}
	intake.store.Set(id, result)

	return result
}

func (intake *JobIntake) Submit(reader io.Reader, source convert.MediaType, originalFileName string) (JobResult, error) {
	id := uuid.New().String()
	for _, found := intake.store.Get(id); found; {
		id = uuid.New().String()
		_, found = intake.store.Get(id)
	}

	filename := id + ".tmp"

	dst, err := intake.fileStore.Create(filename)
	if err != nil {
		return JobResult{}, err
	}
	defer func(dst *os.File) {
		err := dst.Close()
		if err != nil {
			slog.Error("Failed to close file", "err", err)
		}
	}(dst)

	if _, err := dst.ReadFrom(reader); err != nil {
		intake.fileStore.Delete(filename)
		return JobResult{}, err
	}

	return intake.prepare(dst.Name(), id, source, originalFileName), nil
}

func (intake *JobIntake) StartJob(id string, target convert.MediaType) error {
	res, found, swapped := intake.store.CompareAndSwapStatus(id, StatusUploaded, StatusPending)
	if !found {
		return common.NotFoundErr{Msg: "job " + id + " not found"}
	}
	if !swapped {
		return common.IllegalStateErr{Msg: "job not ready"}
	}

	conv, found := intake.registry.Lookup(res.Source, target)
	if !found {
		intake.store.SetStatus(id, StatusUploaded)
		return common.NotFoundErr{Msg: "target converter not found"}
	}

	intake.store.Update(id, func(res JobResult) JobResult {
		res.Target = target
		return res
	})

	job := Job{
		ID:        id,
		FilePath:  res.FilePath,
		Converter: conv,
		Options: convert.Options{
			Target: target,
		},
	}

	if err := intake.queue.Enqueue(job); err != nil {
		intake.store.SetStatus(id, StatusUploaded)
		return err
	}

	return nil
}

func (intake *JobIntake) Lookup(id string) (JobResult, bool) {
	if id == "" {
		return JobResult{}, false
	}

	return intake.store.Get(id)
}

// Delete atomically removes the job's record if its status is deletable, avoiding a race
// with a concurrent StartJob transitioning it out of a deletable status. It returns the
// record as it stood, whether it was found, and whether it was removed.
func (intake *JobIntake) Delete(id string) (JobResult, bool, bool) {
	return intake.store.CompareAndDelete(id)
}
