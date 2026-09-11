package task

import (
	"context"
	"log/slog"
	"sync"
)

type QueueFullErr struct {
}

func (e QueueFullErr) Error() string {
	return "Queue is at full capacity"
}

type Queue struct {
	executor *JobExecutor
	jobs     chan Job
	store    *StatusStore
	workers  int
	wg       sync.WaitGroup
}

func NewQueue(bufferSize, workers int, executor *JobExecutor, store *StatusStore) *Queue {
	return &Queue{
		executor: executor,
		jobs:     make(chan Job, bufferSize),
		store:    store,
		workers:  workers,
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
	name, err := q.executor.Run(ctx, job)
	if err != nil {
		slog.Error("failed to process job", "id", job.ID, "err", err)
		q.store.Set(job.ID, JobResult{JobID: job.ID, Status: StatusFailed, Err: err})
		return
	}

	slog.Info("finished processing job", "id", job.ID)
	q.store.Set(job.ID, JobResult{JobID: job.ID, Status: StatusDone, FilePath: name})
}

func (q *Queue) Enqueue(job Job) error {
	select {
	case q.jobs <- job:
		return nil
	default:
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
