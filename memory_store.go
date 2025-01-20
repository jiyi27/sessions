package sessions

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type MemoryStore struct {
	*baseStore
	mutex      sync.RWMutex
	sessions   map[string]*Session
	gcInterval time.Duration
}

// NewMemoryStore creates and returns a new MemoryStore
// Factory pattern and functional options pattern are used here
func NewMemoryStore(options ...func(store *MemoryStore)) (Store, error) {
	base, err := newBaseStore(defaultOptions(), 16)
	if err != nil {
		return nil, err
	}

	store := &MemoryStore{
		baseStore:  base,
		sessions:   make(map[string]*Session),
		gcInterval: 500 * time.Millisecond,
	}

	// Apply custom options
	for _, op := range options {
		op(store)
	}

	go store.gc()

	return store, nil
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

func (s *MemoryStore) Save(session *Session) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.sessions[session.data.ID] = session
	return nil
}

func (s *MemoryStore) Delete(session *Session) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.sessions, session.data.ID)
	return nil
}

// generateID Generate an unique ID for session
func (s *MemoryStore) generateID() (string, error) {
	// generate an unique random string as session ID
	for {
		if id, err := generateRandomID(s.idLength); err != nil {
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
