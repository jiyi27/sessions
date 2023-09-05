package sessions

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

var mutex sync.RWMutex

func NewSession(name, id string, options Options) *Session {
	return &Session{
		name:    name,
		id:      id,
		isNew:   true,
		expiry:  time.Now().Add(time.Duration(options.MaxAge) * time.Second).Unix(),
		values:  make(map[interface{}]interface{}),
		options: &options,
	}
}

type Session struct {
	name  string
	id    string
	isNew bool
	// expiry is used for deleting expired
	// session internally, user don't need to care.
	expiry  int64
	values  map[interface{}]interface{}
	options *Options
}

// Save saves session into response.
// You should call this function whenever you modify the session.
func (s *Session) Save(w http.ResponseWriter) {
	mutex.RLock()
	opts := *s.options
	mutex.RUnlock()
	http.SetCookie(w, NewCookie(s.name, s.id, &opts))
}

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

// getExpiry used by MemoryStore for deleting expired session internally.
// Users don't need to care about this function.
func (s *Session) getExpiry() int64 {
	mutex.RLock()
	defer mutex.RUnlock()
	return s.expiry
}

// GetValueByKey returns a value whose key is k in the map.
func (s *Session) GetValueByKey(k interface{}) (interface{}, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	v, ok := s.values[k]
	if !ok {
		return nil, fmt.Errorf("failed to get value, no suchkey:%v", k)
	}
	return v, nil
}

// InsertValue
// You should call this function only when insert a new key into the map.
// Do not use a slice, map or other incomparable types as k.
func (s *Session) InsertValue(k, v interface{}) {
	mutex.RLock()
	defer mutex.RUnlock()
	s.values[k] = v
}

// ModifyValueByKey If the provided key exists in map,
// modify the corresponding value stored.
// Otherwise, return error.
func (s *Session) ModifyValueByKey(k, v interface{}) error {
	mutex.RLock()
	defer mutex.RUnlock()
	if _, ok := s.values[k]; !ok {
		return fmt.Errorf("failed to set value, no such key:%v", k)
	}
	s.values[k] = v
	return nil
}

// GetOptions Return a copy of Options of a Session value.
// In case of data race.
func (s *Session) GetOptions() Options {
	mutex.RLock()
	defer mutex.RUnlock()
	return *s.options
}

// SetMaxAge sets the MaxAge of a session.
func (s *Session) SetMaxAge(seconds int) {
	mutex.Lock()
	defer mutex.Unlock()
	s.options.MaxAge = seconds
	// Set expiresTimestamp for deleting expired session.
	// Users don't need to care expiresTimestamp field of a session.
	s.expiry = time.Now().Add(time.Duration(seconds) * time.Second).Unix()
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
