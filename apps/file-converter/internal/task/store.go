package task

import "sync"

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
// It returns the job's record as it stood before the attempt, whether the
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
	return r, true, true
}

func (s *StatusStore) Get(id string) (JobResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.data[id]
	return r, ok
}
