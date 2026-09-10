package task

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"uuid"

	"trox.dev/file-converter/internal/common"
	"trox.dev/file-converter/internal/convert"
)

type QueueFullErr struct {
}

func (e QueueFullErr) Error() string {
	return "Queue is at full capacity"
}

type Queue struct {
	jobs      chan Job
	store     *StatusStore
	fileStore *FileStore
	registry  *convert.Registry
	workers   int
	wg        sync.WaitGroup
}

func NewQueue(bufferSize, workers int, fileStore *FileStore, registry *convert.Registry) *Queue {
	return &Queue{
		jobs:      make(chan Job, bufferSize),
		store:     NewStatusStore(),
		fileStore: fileStore,
		registry:  registry,
		workers:   workers,
	}
}

func (q *Queue) Start(ctx context.Context) {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(ctx)
	}
}

func (q *Queue) worker(ctx context.Context) {
	defer q.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-q.jobs:
			if !ok {
				return // channel closed, drain complete
			}
			q.process(ctx, job)
		}
	}
}

func (q *Queue) process(ctx context.Context, job Job) {
	q.store.SetStatus(job.ID, StatusProcessing)

	slog.Info("processing job", "id", job.ID)
	name, err := q.doWork(ctx, job)
	if err != nil {
		slog.Error("failed to process job", "id", job.ID, "err", err)
		q.store.Set(job.ID, JobResult{JobID: job.ID, Status: StatusFailed, Err: err})
		if name != "" {
			q.fileStore.Delete(name)
		}
		return
	}

	slog.Info("finished processing job", "id", job.ID)
	q.store.Set(job.ID, JobResult{JobID: job.ID, Status: StatusDone, FilePath: name})
}

func (q *Queue) Prepare(path string, src convert.MediaType) JobResult {
	id := uuid.New().String()
	for _, found := q.store.Get(id); found; {
		id = uuid.New().String()
		_, found = q.store.Get(id)
	}

	result := JobResult{
		JobID:    id,
		Status:   StatusUploaded,
		FilePath: path,
		source:   src,
	}
	q.store.Set(id, result)

	return result
}

func (q *Queue) Lookup(id string) (JobResult, bool) {
	return q.store.Get(id)
}

func (q *Queue) StartJob(id string, target convert.MediaType) error {
	res, found, swapped := q.store.CompareAndSwapStatus(id, StatusUploaded, StatusPending)
	if !found {
		return common.NotFoundErr{Msg: "job " + id + " not found"}
	}
	if !swapped {
		return common.IllegalStateErr{Msg: "job not ready"}
	}

	conv, found := q.registry.Lookup(res.source, target)
	if !found {
		q.store.SetStatus(id, StatusUploaded)
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

	select {
	case q.jobs <- job:
		return nil
	default:
		q.store.SetStatus(id, StatusUploaded)
		return QueueFullErr{}
	}
}

func (q *Queue) Shutdown(ctx context.Context) error {
	close(q.jobs)

	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *Queue) doWork(ctx context.Context, job Job) (string, error) {
	handle, err := os.Open(job.FilePath)
	if err != nil {
		return "", err
	}
	defer func(handle *os.File) {
		err := handle.Close()
		if err != nil {
			slog.Warn("failed to close file", "err", err)
		}
	}(handle)

	out, err := q.fileStore.Create(job.ID)
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			slog.Warn("failed to close file", "err", err)
		}
	}(out)

	if err != nil {
		return "", err
	}

	return out.Name(), job.Converter.Convert(ctx, handle, out, job.Options)
}
