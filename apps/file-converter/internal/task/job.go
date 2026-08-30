package task

import "trox.dev/file-converter/internal/convert"

type Job struct {
	ID        string
	FilePath  string
	Opts      convert.Options
	Converter convert.Converter
}

type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusDone       JobStatus = "done"
	StatusFailed     JobStatus = "failed"
)

type JobResult struct {
	JobID  string
	Status JobStatus
	Error  error
}
