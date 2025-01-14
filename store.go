package sessions

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type MemoryStore struct {
	mutex      sync.RWMutex
	sessions   map[string]*Session
	options    *Options
	gcInterval time.Duration
	idLength   int
}

// NewMemoryStore creates and returns a new MemoryStore
// Factory pattern and functional options pattern are used here
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

	// Apply custom options
	for _, op := range options {
		op(s)
	}

	go s.gc()
	return s
}

// Option functions for customizing MemoryStore

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
		s.mutex.RLock()
		session, ok := s.sessions[c.Value]
		s.mutex.RUnlock()
		if ok {
			session.data.IsNew = false
			return session, nil
		}
	}
	// cookie doesn't exist or no corresponding session stored in MemoryStore
	// generate a new session.
	return s.New(name)
}

// New Returns a new session and saves it into underlying store
func (s *MemoryStore) New(name string) (*Session, error) {
	id, err := s.generateID()
	if err != nil {
		return nil, err
	}
	session := NewSession(name, id, *s.options)
	// saves session into underlying store
	s.mutex.Lock()
	s.sessions[session.data.ID] = session
	s.mutex.Unlock()
	return session, nil
}

// generateID Generate an unique ID for session
func (s *MemoryStore) generateID() (string, error) {
	// generate an unique random string as session ID
	for {
		if id, err := GenerateRandomString(s.idLength); err != nil {
			return "", err
		} else {
			s.mutex.RLock()
			_, ok := s.sessions[id]
			s.mutex.RUnlock()
			if !ok {
				return id, nil
			}
		}
	}
}

// gc periodically removes expired sessions
func (s *MemoryStore) gc() {
	ticker := time.NewTicker(s.gcInterval)
	defer ticker.Stop()
	for range ticker.C {
		s.mutex.Lock()
		for k, session := range s.sessions {
			if session.data.Expiry <= time.Now().Unix() {
				delete(s.sessions, k)
			}
		}
		s.mutex.Unlock()
	}
}
