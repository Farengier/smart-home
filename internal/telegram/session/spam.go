package session

import (
	"sync"
	"time"
)

type spam struct {
	times map[int]time.Time
	mtx   sync.Mutex
}

func (s *spam) Get(l int) time.Time {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	t, ok := s.times[l]
	if !ok {
		return time.Time{}
	}
	return t
}

func (s *spam) Set(l int, t time.Time) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.times[l] = t
}
