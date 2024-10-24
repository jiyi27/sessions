package sessions

import (
	"fmt"
	"net/http"
	"time"
	"sync"
)

type MemoryStore struct {
	mu         sync.RWMutex
	sessions   map[string]*Session
	options    *Options
	gcInterval time.Duration
	idLength   int
	// Only use these two channels when gc with expired session set
	ExpiredSession    chan []*Session
	ExpiredSessionErr chan error
}

func NewMemoryStore(options ...func(store *MemoryStore)) *MemoryStore {
	s := &MemoryStore{
		sessions: make(map[string]*Session),
		options: &Options{
			Path:     "/",
			MaxAge:   60,
			SameSite: http.SameSiteDefaultMode,
		},
		gcInterval: time.Millisecond * 500,
		idLength:   16,
	}

	// Apply custom options first
	for _, op := range options {
		op(s)
	}

	// Start default gc if ExpiredSession channel is not initialized
	if s.ExpiredSession == nil {
		go s.gc()
	}

	return s
}

// WithExpiredSessionTracking enables tracking of expired sessions through channels
func WithExpiredSessionTracking() func(*MemoryStore) {
	return func(s *MemoryStore) {
		s.ExpiredSession = make(chan []*Session, 1)
		s.ExpiredSessionErr = make(chan error, 1)
		go s.gcWithTracking()
	}
}

func WithIDLength(l int) func(*MemoryStore) {
	return func(s *MemoryStore) {
		s.idLength = l
	}
}

func WithMaxAge(maxAge int) func(*MemoryStore) {
	return func(s *MemoryStore) {
		s.options.MaxAge = maxAge
	}
}

func WithGCInterval(interval time.Duration) func(*MemoryStore) {
	return func(s *MemoryStore) {
		s.gcInterval = interval
	}
}

// Get returns a session if exists, if it doesn't exist, create a new one.
func (s *MemoryStore) Get(r *http.Request, name string) (*Session, error) {
	if !isCookieNameValid(name) {
		return nil, fmt.Errorf("sessions: invalid character in cookie name: %s", name)
	}
	if c, err := r.Cookie(name); err == nil {
		// check if there is a corresponding session in MemoryStore.
		s.mu.RLock()
		session, ok := s.sessions[c.Value]
		s.mu.RUnlock()
		if ok {
			session.isNew = false
			return session, nil
		}
	}
	// cookie doesn't exist or no corresponding session stored in MemoryStore
	// generate a new session.
	return s.New(name)
}

// New Returns a new session and saves it into underlying store.
func (s *MemoryStore) New(name string) (*Session, error) {
	id, err := s.generateID()
	if err != nil {
		return nil, err
	}
	session := NewSession(name, id, *s.options)
	// saves session into underlying store.
	s.mu.Lock()
	s.sessions[session.id] = session
	s.mu.Unlock()
	return session, nil
}

// generateID Generate an unique ID for session.
func (s *MemoryStore) generateID() (string, error) {
	for {
		if id, err := GenerateRandomString(s.idLength); err != nil {
			return "", err
		} else {
			s.mu.RLock()
			_, ok := s.sessions[id]
			s.mu.RUnlock()
			if !ok {
				return id, nil
			}
		}
	}
}

// Adopted from: https://github.com/gofiber/storage/blob/main/memory/memory.go
func (s *MemoryStore) gc() {
	ticker := time.NewTicker(s.gcInterval)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		for k, session := range s.sessions {
			if session.expiry <= time.Now().Unix() {
				delete(s.sessions, k)
			}
		}
		s.mu.Unlock()
	}
}

func (s *MemoryStore) gcWithTracking() {
	ticker := time.NewTicker(s.gcInterval)
	defer ticker.Stop()
	expired := make([]*Session, 0)
	for range ticker.C {
		// Drop useless elements in last round.
		expired = expired[:0]
		s.mu.Lock()
		for _, session := range s.sessions {
			if session.expiry <= time.Now().Unix() {
				// Send copied expired session to user in case of data race
				copied, err := copySession(session)
				if err != nil {
					s.ExpiredSessionErr <- err
					continue
				}
				expired = append(expired, copied)
				delete(s.sessions, session.id)
			}
		}
		s.mu.Unlock()
		s.ExpiredSession <- expired
	}
}

func copySession(session *Session) (*Session, error) {
	var err error
	cs := NewSession(session.name, session.id, *session.options)
	cs.values, err = DeepCopyMap(session.values)
	if err != nil {
		return nil, fmt.Errorf("failed to copy session %v: %v", session.id, err)
	}
	return cs, nil
}
