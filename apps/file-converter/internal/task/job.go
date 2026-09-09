package task

import "trox.dev/file-converter/internal/convert"

type Job struct {
	ID        string
	FilePath  string
	Converter convert.Converter
	Options   convert.Options
}

type JobStatus string

const (
	StatusUploaded   JobStatus = "uploaded"
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusDone       JobStatus = "done"
	StatusFailed     JobStatus = "failed"
)

type JobResult struct {
	JobID    string
	FilePath string
	Status   JobStatus
	Error    error
	source   convert.MediaType
}
