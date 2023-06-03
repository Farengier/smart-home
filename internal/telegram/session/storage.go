package session

import (
	"sync"
	"time"
)

type Storage struct {
	users    map[int64]*User
	sessions map[int64]*Session
	mtx      sync.RWMutex
}

func New() *Storage {
	a := &Storage{
		users:    make(map[int64]*User),
		sessions: make(map[int64]*Session),
	}
	return a
}

func (a *Storage) Session(chat int64) *Session {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	s, ok := a.sessions[chat]
	if !ok {
		s = &Session{
			ChatID: chat,
			Spam:   &spam{times: map[int]time.Time{}},
		}
		a.sessions[chat] = s
	}
	return s
}

func (a *Storage) LogIn(chat int64, u *User) {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.users[chat] = u
}

func (a *Storage) User(chat int64) (*User, bool) {
	a.mtx.RLock()
	defer a.mtx.RUnlock()
	u, ok := a.users[chat]
	return u, ok
}
