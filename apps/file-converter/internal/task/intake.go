package task

import (
	"io"
	"log/slog"
	"mime/multipart"
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

func (intake *JobIntake) prepare(path, id string, src convert.MediaType) JobResult {
	result := JobResult{
		JobID:    id,
		Status:   StatusUploaded,
		FilePath: path,
		source:   src,
	}
	intake.store.Set(id, result)

	return result
}

func (intake *JobIntake) Submit(reader io.Reader, source convert.MediaType) (JobResult, error) {
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
	defer func(dst multipart.File) {
		err := dst.Close()
		if err != nil {
			slog.Error("Failed to close file", "err", err)
		}
	}(dst)

	if _, err := dst.ReadFrom(reader); err != nil {
		intake.fileStore.Delete(filename)
		return JobResult{}, err
	}

	return intake.prepare(dst.Name(), id, source), nil
}

func (intake *JobIntake) StartJob(id string, target convert.MediaType) error {
	res, found, swapped := intake.store.CompareAndSwapStatus(id, StatusUploaded, StatusPending)
	if !found {
		return common.NotFoundErr{Msg: "job " + id + " not found"}
	}
	if !swapped {
		return common.IllegalStateErr{Msg: "job not ready"}
	}

	conv, found := intake.registry.Lookup(res.source, target)
	if !found {
		intake.store.SetStatus(id, StatusUploaded)
		return common.NotFoundErr{Msg: "target converter not found"}
	}

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
