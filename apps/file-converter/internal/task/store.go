package task

import (
	"sync"
)

type StatusStore struct {
	mu   sync.RWMutex
	data map[string]JobResult
}

func NewStatusStore() *StatusStore {
	return &StatusStore{
		data: make(map[string]JobResult),
	}
}

func (s *StatusStore) Set(id string, r JobResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = r
}

// Update updates the job result for the given id. If the job is not found, nothing happens.
func (s *StatusStore) Update(id string, fn func(JobResult) JobResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.data[id]
	if !ok {
		return
	}

	s.data[id] = fn(r)
}

func (s *StatusStore) SetStatus(id string, status JobStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.data[id]
	if !ok {
		return
	}

	r.Status = status
	s.data[id] = r
}

// CompareAndSwapStatus atomically moves the job's status from `from` to `to`.
// It returns the job's updated record if the swap was successful or the old one if it was unsuccessful, whether the
// job was found, and whether the swap happened (i.e. its status was `from`).
func (s *StatusStore) CompareAndSwapStatus(id string, from, to JobStatus) (result JobResult, found, swapped bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.data[id]
	if !ok {
		return JobResult{}, false, false
	}

	if r.Status != from {
		return r, true, false
	}

	updated := r
	updated.Status = to
	s.data[id] = updated
	return updated, true, true
}

// CompareAndDelete removes the record for id if it exists and its status is deletable.
// It returns the record as it stood, whether it was found, and whether it was removed.
func (s *StatusStore) CompareAndDelete(id string) (result JobResult, found, deleted bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.data[id]
	if !ok {
		return JobResult{}, false, false
	}
	if !r.Status.Deletable() {
		return r, true, false
	}
	delete(s.data, id)
	return r, true, true
}

func (s *StatusStore) Get(id string) (JobResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.data[id]
	return r, ok
}
