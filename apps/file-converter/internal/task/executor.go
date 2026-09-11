package task

import (
	"context"
	"log/slog"
	"os"
)

type JobExecutor struct {
	fileStore *FileStore
}

func NewJobExecutor(fileStore *FileStore) *JobExecutor {
	return &JobExecutor{
		fileStore: fileStore,
	}
}

func (exec *JobExecutor) Run(ctx context.Context, job Job) (string, error) {
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

	out, err := exec.fileStore.Create(job.ID)
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			slog.Warn("failed to close file", "err", err)
		}
	}(out)

	if err != nil {
		return "", err
	}

	if err := job.Converter.Convert(ctx, handle, out, job.Options); err != nil {
		exec.fileStore.Delete(out.Name())
		return "", err
	}

	return out.Name(), nil
}
