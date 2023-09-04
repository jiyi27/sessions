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
	// GetAllSessions() ([]MySession, error)

	// New should create and return a new session.
	//
	// Note that New should never return a nil session, even in the case of
	// an error if using the Registry infrastructure to cache the session.
	New(r *http.Request, name string) (*MySession, error)

	// Save should persist session to the underlying store implementation.
	Save(r *http.Request, w http.ResponseWriter, s *MySession) error
}

// memoryStore ------------------------------------------------------------

// mutexes are frequently wrapped up in a `struct` with the value they control.
// we should put memoryMutex as a field of memoryStore
// struct MySession has a field which is Store
// However, we should not copy a `sync.Mutex` value as that breaks the invariants of the mutex.
// Therefore, for aviod copying, we put memoryMutex as a package level variable.
// https://dave.cheney.net/2016/03/19/should-methods-be-declared-on-t-or-t
var memoryMutex sync.RWMutex

func newMemoryStore() *memoryStore {
	s := memoryStore{
		sessions: map[string]*sessionInfo{},
		// default settings
		Options: &Options{
			Path:     "/",
			MaxAge:   60,
			SameSite: http.SameSiteDefaultMode,
		},
	}
	go s.monitorExpiredSessions()
	return &s
}

// not thread-safe
// each request will have one or more goroutines
type memoryStore struct {
	// reason that saves a pointer to MySession rather the value of MySession here:
	// https://stackoverflow.com/a/29868656/16317008
	sessions map[string]*sessionInfo
	Options  *Options
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
	session := NewMySession(name, id, s)
	*session.Options = *s.Options
	// cookie found
	if c, errCookie := r.Cookie(name); errCookie == nil {
		// cookie is correct
		memoryMutex.RLock()
		defer memoryMutex.RUnlock()
		sInfo, ok := s.sessions[c.Value]
		if ok {
			// deep copy value here, prevent data race
			session.Values, err = DeepCopyMap(sInfo.session.Values)
			if err != nil {
				return nil, err
			}
			*session.Options = *sInfo.session.Options
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
	sessionInfoPtr := &sessionInfo{
		session:          session,
		expiresTimestamp: time.Now().Add(d).Unix(),
	}
	memoryMutex.Lock()
	s.sessions[session.id] = sessionInfoPtr
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

func (s *memoryStore) monitorExpiredSessions() {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		if s.isEmpty() {
			continue
		}
		for k, info := range s.sessions {
			if info.expiresTimestamp >= time.Now().Unix() {
				s.delete(k)
			}
		}
	}
}

func (s *memoryStore) delete(k string) {
	// check the concurrency part: https://go.dev/blog/maps
	// mutex: https://stackoverflow.com/a/19168242/16317008
	memoryMutex.Lock()
	defer memoryMutex.Unlock()
	delete(s.sessions, k)
}

func (s *memoryStore) isEmpty() bool {
	memoryMutex.RLock()
	defer memoryMutex.RUnlock()
	return len(s.sessions) == 0
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
	}
	go s.monitorExpiredSessions()
	return &s
}

type cookieStore struct {
	sessions map[string]*Session
	options  *Options
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
		session.SetIsNew(false)
		mutex.RUnlock()
		if ok {
			return session, nil
		}
	}
	// cookie doesn't exist or no corresponding session stored in cookieStore
	// generate a new session.
	return s.New(name)
}

// New Return a new session.
func (s *cookieStore) New(name string) (*Session, error) {
	id, err := s.generateID(32)
	if err != nil {
		return nil, err
	}
	session := NewSession(name, id, *s.options)
	return session, nil
}

// Save saves session into response and the underlying store.
func (s *cookieStore) Save(w http.ResponseWriter, session *Session) error {
	opts := session.GetOptions()
	http.SetCookie(w, NewCookie(session.name, session.id, &opts))
	s.save(session)
	return nil
}

func (s *cookieStore) save(session *Session) {
	memoryMutex.Lock()
	s.sessions[session.id] = session
	memoryMutex.Unlock()
}

// generateID Generate an unique ID for session.
func (s *cookieStore) generateID(n int) (string, error) {
	for {
		if id, err := GenerateRandomString(n); err != nil {
			return "", err
		} else {
			memoryMutex.RLock()
			_, ok := s.sessions[id]
			memoryMutex.RUnlock()
			if !ok {
				return id, nil
			}
		}
	}
}

func (s *cookieStore) monitorExpiredSessions() {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		if s.isEmpty() {
			continue
		}
		for k, info := range s.sessions {
			if info.expiresTimestamp >= time.Now().Unix() {
				s.delete(k)
			}
		}
	}
}

func (s *cookieStore) delete(k string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(s.sessions, k)
}

func (s *cookieStore) isEmpty() bool {
	mutex.RLock()
	defer mutex.RUnlock()
	return len(s.sessions) == 0
}
