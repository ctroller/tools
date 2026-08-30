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

func (s *StatusStore) Get(id string) (JobResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.data[id]
	return r, ok
}
