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
	// StatusUploaded indicates that a job has been uploaded and is ready to be processed.
	StatusUploaded JobStatus = "uploaded"
	// StatusPending indicates that a job is waiting to be processed.
	StatusPending JobStatus = "pending"
	// StatusProcessing indicates that a job is currently being processed.
	StatusProcessing JobStatus = "processing"
	// StatusDone indicates that a job has completed successfully.
	StatusDone JobStatus = "done"
	// StatusFailed indicates that a job has encountered an error and did not complete successfully.
	StatusFailed JobStatus = "failed"
)

// Done returns if the given status is done (StatusDone || StatusFailed)
func (s JobStatus) Done() bool {
	return s == StatusDone || s == StatusFailed
}

type JobResult struct {
	JobID    string
	FilePath string
	Status   JobStatus
	Err      error
	source   convert.MediaType
}

func (r JobResult) Error() string {
	if r.Err != nil {
		return r.Err.Error()
	}
	return ""
}
