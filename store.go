package sessions

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Store Interface ------------------------------------------------------------

// Store is an interface for custom session stores.
type Store interface {
	// Get should return a session if exists, if it doesn't exist, create a new one
	// If Get doesn't create a new one and return nil instead, the user call this
	// function have to create a session eventually, but creating a session should
	// not let user do, because they don't have to know the complicated thing
	Get(r *http.Request, name string) (*MySession, error)

	// Should not have GetAllSessions method,
	// first, user should not consider concurrency issue
	// second, if return a copy of all sessions, the cost is huge.
	// GetAllSessions should return all sessions in a slice
	// GetAllSessions() ([]Session, error)

	// New should create and return a new session.
	//
	// Note that New should never return a nil session, even in the case of
	// an error if using the Registry infrastructure to cache the session.
	New(r *http.Request, name string) (*MySession, error)

	// Save should persist session to the underlying store implementation.
	Save(r *http.Request, w http.ResponseWriter, s *MySession) error
}

// cookieStore same as memoryStore but no deep-copy --------------------------------------------------

func newCookieStore() *cookieStore {
	s := cookieStore{
		sessions: make(map[string]*Session),
		options: &Options{
			Path:     "/",
			MaxAge:   60,
			SameSite: http.SameSiteDefaultMode,
		},
		gcInterval: time.Millisecond * 500,
	}
	go s.gc()
	return &s
}

type cookieStore struct {
	sessions   map[string]*Session
	options    *Options
	gcInterval time.Duration
}

// Get returns a session if exists, if it doesn't exist, create a new one.
func (s *cookieStore) Get(r *http.Request, name string) (*Session, error) {
	if !isCookieNameValid(name) {
		return nil, fmt.Errorf("sessions: invalid character in cookie name: %s", name)
	}
	if c, err := r.Cookie(name); err == nil {
		// if cookie exists in the request
		// and check if there is a corresponding session in cookieStore.
		mutex.RLock()
		session, ok := s.sessions[c.Value]
		mutex.RUnlock()
		if ok {
			session.SetIsNew(false)
			return session, nil
		}
	}
	// cookie doesn't exist or no corresponding session stored in cookieStore
	// generate a new session.
	return s.New(name)
}

// New Returns a new session and saves it into underlying store.
func (s *cookieStore) New(name string) (*Session, error) {
	id, err := s.generateID(32)
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
func (s *cookieStore) generateID(n int) (string, error) {
	for {
		if id, err := GenerateRandomString(n); err != nil {
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
func (s *cookieStore) gc() {
	ticker := time.NewTicker(s.gcInterval)
	defer ticker.Stop()
	var expired []string
	for range ticker.C {
		if s.isEmpty() {
			continue
		}
		mutex.RLock()
		for k, session := range s.sessions {
			if session.expiry <= time.Now().Unix() {
				expired = append(expired, k)
			}
		}
		mutex.RUnlock()
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

func (s *cookieStore) isEmpty() bool {
	mutex.RLock()
	defer mutex.RUnlock()
	return len(s.sessions) == 0
}

// memoryStore ------------------------------------------------------------

// mutexes are frequently wrapped up in a `struct` with the value they control.
// we should put memoryMutex as a field of memoryStore
// struct Session has a field which is Store
// However, we should not copy a `sync.Mutex` value as that breaks the invariants of the mutex.
// Therefore, for avoid copying, we put memoryMutex as a package level variable.
// https://dave.cheney.net/2016/03/19/should-methods-be-declared-on-t-or-t
var memoryMutex sync.RWMutex

func newMemoryStore() *memoryStore {
	s := memoryStore{
		sessions: map[string]*MySession{},
		// default settings
		Options: &Options{
			Path:     "/",
			MaxAge:   60,
			SameSite: http.SameSiteDefaultMode,
		},
		gcInterval: time.Millisecond * 500,
	}
	go s.gc()
	return &s
}

// not thread-safe
// each request will have one or more goroutines
type memoryStore struct {
	// reason that saves a pointer to Session rather the value of Session here:
	// https://stackoverflow.com/a/29868656/16317008
	sessions   map[string]*MySession
	Options    *Options
	gcInterval time.Duration
}

// Get always return a new session, if the session corresponding to the cookie exists,
// copy the session info into new session for preventing data race.
// Because there is more than one goroutine need to access memoryStore.
// Sessions are stored in memoryStore, if we don't make a copy here, users need to lock almost
// everything when trying to read or change the session.
func (s *memoryStore) Get(r *http.Request, name string) (*MySession, error) {
	if !isCookieNameValid(name) {
		return nil, fmt.Errorf("sessions: invalid character in cookie name: %s", name)
	}
	return s.New(r, name)
}

// New Return a new session.
func (s *memoryStore) New(r *http.Request, name string) (*MySession, error) {
	id, err := s.generateID(32)
	if err != nil {
		return nil, err
	}
	session := NewMySession(name, id, s.Options.MaxAge, s)
	*session.Options = *s.Options
	// cookie found
	if c, errCookie := r.Cookie(name); errCookie == nil {
		// cookie is correct
		memoryMutex.RLock()
		defer memoryMutex.RUnlock()
		se, ok := s.sessions[c.Value]
		if ok {
			// deep copy value here, prevent data race
			session.Values, err = DeepCopyMap(se.Values)
			if err != nil {
				return nil, err
			}
			*session.Options = *se.Options
			session.id = c.Value
			session.IsNew = false
		}
	}
	return session, nil
}

// Save saves session into response and the underlying store.
func (s *memoryStore) Save(_ *http.Request, w http.ResponseWriter,
	session *MySession) error {
	// if session expires, set cookie value = ""
	http.SetCookie(w, NewCookie(session.name, session.id, session.Options))
	s.save(session)
	return nil
}

func (s *memoryStore) save(session *MySession) {
	d := time.Duration(session.Options.MaxAge) * time.Second
	session.expiry = time.Now().Add(d).Unix()
	memoryMutex.Lock()
	s.sessions[session.id] = session
	memoryMutex.Unlock()
}

// generateID Generate an unique ID for session.
func (s *memoryStore) generateID(n int) (string, error) {
	for {
		id, err := GenerateRandomString(n)
		if err != nil {
			return "", err
		}
		memoryMutex.RLock()
		_, ok := s.sessions[id]
		memoryMutex.RUnlock()
		if !ok {
			return id, nil
		}
	}
}

func (s *memoryStore) gc() {
	ticker := time.NewTicker(s.gcInterval)
	defer ticker.Stop()
	var expired []string
	// check the concurrency part: https://go.dev/blog/maps
	// mutex: https://stackoverflow.com/a/19168242/16317008
	for range ticker.C {
		if s.isEmpty() {
			continue
		}
		memoryMutex.RLock()
		for k, session := range s.sessions {
			if session.expiry <= time.Now().Unix() {
				expired = append(expired, k)
			}
		}
		memoryMutex.RUnlock()
		memoryMutex.Lock()
		// Double-checked locking.
		// User might have reset the max age of the session in the meantime.
		for i := range expired {
			v := s.sessions[expired[i]]
			if v.expiry <= time.Now().Unix() {
				delete(s.sessions, expired[i])
			}
		}
		memoryMutex.Unlock()
	}
}

func (s *memoryStore) isEmpty() bool {
	memoryMutex.RLock()
	defer memoryMutex.RUnlock()
	return len(s.sessions) == 0
}
