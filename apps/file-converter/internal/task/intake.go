package task

import (
	"uuid"

	"trox.dev/file-converter/internal/common"
	"trox.dev/file-converter/internal/convert"
)

type JobIntake struct {
	store    *StatusStore
	registry *convert.Registry
	queue    *Queue
}

func NewJobIntake(registry *convert.Registry, queue *Queue, store *StatusStore) *JobIntake {
	return &JobIntake{
		store, registry, queue,
	}
}

func (intake *JobIntake) Prepare(path string, src convert.MediaType) JobResult {
	id := uuid.New().String()
	for _, found := intake.store.Get(id); found; {
		id = uuid.New().String()
		_, found = intake.store.Get(id)
	}

	result := JobResult{
		JobID:    id,
		Status:   StatusUploaded,
		FilePath: path,
		source:   src,
	}
	intake.store.Set(id, result)

	return result
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
	return intake.store.Get(id)
}
