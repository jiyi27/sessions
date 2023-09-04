package sessions

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// NewMySession is called by session stores to create a new session instance.
func NewMySession(name, id string, store Store) *MySession {
	return &MySession{
		name:    name,
		id:      id,
		Values:  make(map[interface{}]interface{}),
		IsNew:   true,
		Options: new(Options),
		store:   store,
	}
}

// MySession stores the values and optional configuration for a session.
type MySession struct {
	name string
	id   string
	// Values contain the user-data for the session.
	// Maps are reference types, so they are always passed by reference.
	// So don't need to save as a pointer here.
	Values  map[interface{}]interface{}
	Options *Options
	IsNew   bool
	store   Store
}

// Save is a convenience method to save this session. It is the same as calling
// store.Save(request, response, session). You should call Save before writing to
// the response or returning from the handler.
func (s *MySession) Save(r *http.Request, w http.ResponseWriter) error {
	return s.store.Save(r, w, s)
}

// Name returns the name used to register the session.
func (s *MySession) Name() string {
	return s.name
}

// Store returns the session store used to register the session.
func (s *MySession) Store() Store {
	return s.store
}

type sessionInfo struct {
	session          *MySession
	expiresTimestamp int64
}

// MySession --------------------------------------------------------------------------

var mutex sync.RWMutex

func NewSession(name, id string, options Options) *Session {
	return &Session{
		name:             name,
		id:               id,
		isNew:            true,
		expiresTimestamp: time.Now().Add(time.Duration(options.MaxAge)).Unix(),
		values:           make(map[interface{}]interface{}),
		options:          &options,
	}
}

type Session struct {
	name  string
	id    string
	isNew bool
	// expiresTimestamp is used for deleting expired
	// session internally, user don't need to care.
	expiresTimestamp int64
	values           map[interface{}]interface{}
	options          *Options
}

// Infos ---------------------------------------

func (s *Session) GetID() string {
	mutex.RLock()
	defer mutex.RUnlock()
	return s.id
}

func (s *Session) GetName() string {
	mutex.RLock()
	defer mutex.RUnlock()
	return s.name
}

func (s *Session) IsNew() bool {
	mutex.RLock()
	defer mutex.RUnlock()
	return s.isNew
}

func (s *Session) SetIsNew(isNew bool) {
	mutex.Lock()
	defer mutex.Unlock()
	s.isNew = isNew
}

// getExpiresTimestamp used by cookieStore for deleting expired session internally
func (s *Session) getExpiresTimestamp() int64 {
	mutex.RLock()
	defer mutex.RUnlock()
	return s.expiresTimestamp
}

// Values ---------------------------------------

// GetValue Do not use function, slice or other incomparable types as k.
func (s *Session) GetValue(k interface{}) (interface{}, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	v, ok := s.values[k]
	if !ok {
		return nil, fmt.Errorf("failed to get value, no such value with key=%v", k)
	}
	return v, nil
}

// AddValue Do not use function, slice or other incomparable types as k.
func (s *Session) AddValue(k, v interface{}) {
	mutex.RLock()
	defer mutex.RUnlock()
	s.values[k] = v
}

// Options ---------------------------------------

func (s *Session) GetOptions() Options {
	mutex.RLock()
	defer mutex.RUnlock()
	return *s.options
}

// SetMaxAge will set both expiresTimestamp and MaxAge of Options
func (s *Session) SetMaxAge(ma int) {
	mutex.Lock()
	defer mutex.Unlock()
	s.options.MaxAge = ma
	s.expiresTimestamp = time.Now().Add(time.Duration(ma)).Unix()
}

func (s *Session) SetCookiePath(path string) {
	mutex.Lock()
	defer mutex.Unlock()
	s.options.Path = path
}

func (s *Session) SetCookieDomain(domain string) {
	mutex.Lock()
	defer mutex.Unlock()
	s.options.Domain = domain
}

func (s *Session) SetCookieSecure(secure bool) {
	mutex.Lock()
	defer mutex.Unlock()
	s.options.Secure = secure
}

func (s *Session) SetCookieHttpOnly(ho bool) {
	mutex.Lock()
	defer mutex.Unlock()
	s.options.HttpOnly = ho
}

func (s *Session) SetCookieSameSite(ss http.SameSite) {
	mutex.Lock()
	defer mutex.Unlock()
	s.options.SameSite = ss
}

func (s *Session) GetMaxAge() int {
	mutex.RLock()
	defer mutex.RUnlock()
	return s.options.MaxAge
}

func (s *Session) GetCookiePath() string {
	mutex.RLock()
	defer mutex.RUnlock()
	return s.options.Path
}

func (s *Session) GetCookieDomain() string {
	mutex.RLock()
	defer mutex.RUnlock()
	return s.options.Domain
}

func (s *Session) GetCookieSecure() bool {
	mutex.RLock()
	defer mutex.RUnlock()
	return s.options.Secure
}

func (s *Session) GetCookieHttpOnly() bool {
	mutex.RLock()
	defer mutex.RUnlock()
	return s.options.HttpOnly
}

func (s *Session) GetCookieSameSite() http.SameSite {
	mutex.RLock()
	defer mutex.RUnlock()
	return s.options.SameSite
}
