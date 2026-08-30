package task

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
)

type Queue struct {
	jobs    chan Job
	store   *StatusStore
	workers int
	wg      sync.WaitGroup
}

func NewQueue(bufferSize, workers int, store *StatusStore) *Queue {
	return &Queue{
		jobs:    make(chan Job, bufferSize),
		store:   store,
		workers: workers,
	}
}

func (q *Queue) Start(ctx context.Context) {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(ctx, i)
	}
}

func (q *Queue) worker(ctx context.Context, id int) {
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
	q.store.Set(job.ID, JobResult{Status: StatusProcessing})

	slog.Info("processing job", "id", job.ID)
	if err := doWork(ctx, job); err != nil {
		slog.Error("failed to process job", "id", job.ID, "err", err)
		q.store.Set(job.ID, JobResult{Status: StatusFailed, Error: err})
		return
	}

	slog.Info("finished processing job", "id", job.ID)
	q.store.Set(job.ID, JobResult{Status: StatusDone})
}

func (q *Queue) Submit(job Job) error {
	q.store.Set(job.ID, JobResult{Status: StatusPending})
	select {
	case q.jobs <- job:
		return nil
	default:
		return errors.New("queue full") // non-blocking submit
	}
}

func (q *Queue) Shutdown() {
	close(q.jobs)
	q.wg.Wait()
}

func doWork(ctx context.Context, job Job) error {
	handle, err := os.Open(job.FilePath)
	if err != nil {
		return err
	}
	defer func(handle *os.File) {
		err := handle.Close()
		if err != nil {
			slog.Warn("failed to close file", "err", err)
		}
	}(handle)

	out, err := os.CreateTemp("/tmp", job.FilePath+".converted")
	if out == nil {
		return errors.New("failed to create temp file")
	}
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			slog.Warn("failed to close file", "err", err)
		}
	}(out)

	return job.Converter.Convert(ctx, handle, out, job.Opts)
}
