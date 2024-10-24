package sessions

import (
	"net/http"
	"sync"
	"time"
)

func NewSession(name, id string, options Options) *Session {
	return &Session{
		name:    name,
		id:      id,
		isNew:   true,
		expiry:  time.Now().Add(time.Duration(options.MaxAge) * time.Second).Unix(),
		values:  make(map[string]interface{}),
		options: &options,
	}
}

type Session struct {
	mu      sync.RWMutex
	name  string
	id    string
	isNew bool
	// expiry is used for deleting expired session internally
	expiry  int64
	values  map[string]interface{}
	options *Options
}

// Save saves session into response.
// You should call this function whenever you modify the session.
func (s *Session) Save(w http.ResponseWriter) {
	s.mu.RLock()
	opts := *s.options
	s.mu.RUnlock()
	http.SetCookie(w, NewCookie(s.name, s.id, &opts))
}

func (s *Session) GetID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.id
}

func (s *Session) GetName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.name
}

func (s *Session) IsNew() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isNew
}

func (s *Session) SetIsNew(isNew bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isNew = isNew
}

// // getExpiry used by MemoryStore for deleting expired session internally.
// // Users don't need to care about this function.
// func (s *Session) getExpiry() int64 {
// 	mutex.RLock()
// 	defer mutex.RUnlock()
// 	return s.expiry
// }

// GetValueByKey returns a value whose key is k in the map.
func (s *Session) GetValueByKey(k string) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.values[k]
}

// SetValue
// You should call this function only when insert a new key into the map.
// Do not use a slice, map or other incomparable types as k.
func (s *Session) SetValue(k string, v interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[k] = v
}

// GetOptions Return a copy of Options of a Session value.
// In case of data race.
func (s *Session) GetOptions() Options {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.options
}

// SetMaxAge sets the MaxAge of a session.
func (s *Session) SetMaxAge(seconds int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.options.MaxAge = seconds
	// Set expiresTimestamp for deleting expired session.
	// Users don't need to care expiresTimestamp field of a session.
	s.expiry = time.Now().Add(time.Duration(seconds) * time.Second).Unix()
}

func (s *Session) SetCookiePath(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.options.Path = path
}

func (s *Session) SetCookieDomain(domain string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.options.Domain = domain
}

func (s *Session) SetCookieSecure(secure bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.options.Secure = secure
}

func (s *Session) SetCookieHttpOnly(ho bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.options.HttpOnly = ho
}

func (s *Session) SetCookieSameSite(ss http.SameSite) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.options.SameSite = ss
}

func (s *Session) GetMaxAge() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.options.MaxAge
}

func (s *Session) GetCookiePath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.options.Path
}

func (s *Session) GetCookieDomain() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.options.Domain
}

func (s *Session) GetCookieSecure() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.options.Secure
}

func (s *Session) GetCookieHttpOnly() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.options.HttpOnly
}

func (s *Session) GetCookieSameSite() http.SameSite {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.options.SameSite
}
