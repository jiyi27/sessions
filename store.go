package sessions

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// Store Interface ------------------------------------------------------------

// Store is an interface for custom session stores.
//
// See CookieStore for example.
type Store interface {
	// Get should return a session if exists, if it doesn't exist create a new one
	// If Get doesn't create a new one and return nil instead, the user call this
	// function have to create a session eventually, but creating a session should
	// not let user do, because they don't have to know the complicated thing
	Get(r *http.Request, name string) (*Session, error)

	// GetAllSessions should return all sessions in a slice
	// copy-on-write for concurrency
	// or read-only
	GetAllSessions() ([]Session, error)

	// New should create and return a new session.
	//
	// Note that New should never return a nil session, even in the case of
	// an error if using the Registry infrastructure to cache the session.
	New(r *http.Request, name string) (*Session, error)

	// Save should persist session to the underlying store implementation.
	Save(r *http.Request, w http.ResponseWriter, s *Session) error
}

// memoryStore ------------------------------------------------------------

func newMemoryStore() *memoryStore {
	return &memoryStore{
		sessions: map[string]*sessionInfo{},
		// default settings
		options: &Options{
			Path:     "/",
			MaxAge:   60,
			SameSite: http.SameSiteDefaultMode,
		},
		mutex: sync.RWMutex{},
	}
}

// not thread-safe
// each request will have one or more goroutines
type memoryStore struct {
	// reason that saves a pointer to Session rather the value of Session here:
	// https://stackoverflow.com/a/29868656/16317008
	sessions map[string]*sessionInfo
	options  *Options
	mutex    sync.RWMutex
}

// Get return a session, if session doesn't exist create a new one
func (s *memoryStore) Get(r *http.Request, name string) (*Session, error) {
	// return ErrNoCookie if not found
	if c, err := r.Cookie(name); err != nil {
		session, _ := s.New(name)
		return session, nil
	} else {
		// IsNew = false

		return nil, err
	}
}

// New Return a new session and save it to memoryStore
func (s *memoryStore) New(name string) (*Session, error) {
	if id, err := GenerateRandomString(16); err != nil {
		// error handling
		return nil, err
	} else {
		session := NewSession(name, id, s.options)
		d := time.Duration(options.MaxAge) * time.Second
		// Check the Concurrency part: https://go.dev/blog/maps
		// mutext: https://stackoverflow.com/a/19168242/16317008
		s.mutex.Lock()
		defer s.mutex.Unlock()
		s.sessions[id] = &sessionInfo{
			session:          session,
			expiresTimestamp: time.Now().Add(d).Unix(),
		}
		return session, nil
	}
}

// Save adds a single session to the response
func (s *memoryStore) Save(r *http.Request, w http.ResponseWriter,
	session *Session) error {
	// if session expires, set cookie value = ""
	http.SetCookie(w, NewCookie(session.name, session.id, session.Options))
	s.save(session)
	return nil
}

func (s *memoryStore) save(session *Session) {

}

func (s *memoryStore) deleteExpiredSessions() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for k, session := range s.sessions {
		d := time.Duration(session.Options.MaxAge) * time.Second
	}

}

// https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb
func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz#@!*&%-_+="
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret = append(ret, letters[num.Int64()])
	}

	return string(ret), nil
}
