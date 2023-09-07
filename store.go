package sessions

import (
	"fmt"
	"net/http"
	"time"
)

// NewMemoryStore returns a new MemoryStore.
// Parameter length is the length of session id,
// you should set it to 32~64, don't make it too big for better performance,
// we save it as a key in a map.
// Parameter enable is optional, default is false.
// Check:
func NewMemoryStore(length int, enable ...bool) *MemoryStore {
	s := MemoryStore{
		sessions: make(map[string]*Session),
		options: &Options{
			Path:     "/",
			MaxAge:   60,
			SameSite: http.SameSiteDefaultMode,
		},
		gcInterval: time.Millisecond * 500,
		idLength:   length,
		ready:      make(chan []string),
		done:       make(chan struct{}),
	}
	if len(enable) > 0 {
		s.enable = true
	}
	go s.gc()
	return &s
}

type MemoryStore struct {
	sessions   map[string]*Session
	options    *Options
	gcInterval time.Duration
	idLength   int
	ready      chan []string
	done       chan struct{}
	enable     bool
}

// Get returns a session if exists, if it doesn't exist, create a new one.
func (s *MemoryStore) Get(r *http.Request, name string) (*Session, error) {
	if !isCookieNameValid(name) {
		return nil, fmt.Errorf("sessions: invalid character in cookie name: %s", name)
	}
	if c, err := r.Cookie(name); err == nil {
		// if cookie exists in the request
		// and check if there is a corresponding session in MemoryStore.
		mutex.RLock()
		session, ok := s.sessions[c.Value]
		mutex.RUnlock()
		if ok {
			session.SetIsNew(false)
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
	mutex.Lock()
	s.sessions[session.id] = session
	mutex.Unlock()
	return session, nil
}

// generateID Generate an unique ID for session.
func (s *MemoryStore) generateID() (string, error) {
	for {
		if id, err := GenerateRandomString(s.idLength); err != nil {
			return "", err
		} else {
			mutex.RLock()
			_, ok := s.sessions[id]
			mutex.RUnlock()
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
	var expired []string
	for range ticker.C {
		if s.isEmpty() {
			continue
		}
		// Drop useless elements in last round.
		expired = expired[:0]
		mutex.RLock()
		for k, session := range s.sessions {
			if session.expiry <= time.Now().Unix() {
				expired = append(expired, k)
			}
		}
		mutex.RUnlock()
		// No session expires.
		if len(expired) == 0 {
			continue
		}
		// Get expired sessions enabled, notify everytime.
		if s.enable {
			// notify & send data to channel
			s.ready <- expired
			// wait for the operation done
			<-s.done
		}
		mutex.Lock()
		// Double-checked locking.
		// User might have reset the max age of the session in the meantime.
		for i := range expired {
			v := s.sessions[expired[i]]
			if v.expiry <= time.Now().Unix() {
				delete(s.sessions, expired[i])
			}
		}
		mutex.Unlock()
	}
}

func (s *MemoryStore) isEmpty() bool {
	mutex.RLock()
	defer mutex.RUnlock()
	return len(s.sessions) == 0
}

func (s *MemoryStore) GetExpiredSessions(expiredSessions chan []*Session, errSession chan error) {
	sessions := make([]*Session, 0)
	// Block until the data is ready.
	for expired := range s.ready {
		for _, id := range expired {
			mutex.RLock()
			session, err := s.copySession(s.sessions[id])
			mutex.RUnlock()
			if err != nil {
				errSession <- err
				continue
			}
			sessions = append(sessions, session)
		}
		// Notify gc
		s.done <- struct{}{}
		expiredSessions <- sessions
	}
}

func (s *MemoryStore) copySession(session *Session) (*Session, error) {
	var err error
	cs := NewSession(session.name, session.id, *session.options)
	cs.values, err = DeepCopyMap(session.values)
	if err != nil {
		return nil, fmt.Errorf("failed to copy session: %v", err)
	}
	return cs, nil
}
