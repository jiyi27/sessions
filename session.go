package sessions

import (
	"net/http"
	"sync"
	"time"
)

type Session struct {
	data  *sessionData
	mutex sync.RWMutex // 零值即可用,不用初始化
}

// sessionData 内部的数据结构, 用于序列化
type sessionData struct {
	Name    string                 `json:"name"`
	ID      string                 `json:"id"`
	IsNew   bool                   `json:"is_new"`
	Expiry  int64                  `json:"expiry"`
	Values  map[string]interface{} `json:"values"`  // sync.Map 对 redis 存储支持不友好, 序列化/反序列化需要额外的转换步骤
	Options *Options               `json:"options"` // cookie 相关配置
}

func NewSession(name, id string, options Options) *Session {
	return &Session{
		data: &sessionData{
			Name:    name,
			ID:      id,
			IsNew:   true,
			Expiry:  time.Now().Add(time.Duration(options.MaxAge) * time.Second).Unix(),
			Values:  make(map[string]interface{}),
			Options: &options,
		},
	}
}

// Save saves session into response.
// You should call this function whenever you modify the session.
func (s *Session) Save(w http.ResponseWriter) {
	s.mutex.RLock()
	opts := *s.data.Options
	s.mutex.RUnlock()
	http.SetCookie(w, NewCookie(s.data.Name, s.data.ID, &opts))
}

func (s *Session) GetID() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.data.ID
}

func (s *Session) GetName() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.data.Name
}

func (s *Session) IsNew() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.data.IsNew
}

func (s *Session) SetIsNew(isNew bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data.IsNew = isNew
}

// GetValueByKey returns a value whose key is k in the map.
func (s *Session) GetValueByKey(k string) interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.data.Values[k]
}

// SetValue
// You should call this function only when insert a new key into the map.
// Do not use a slice, map or other incomparable types as k.
func (s *Session) SetValue(k string, v interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data.Values[k] = v
}

// GetOptions Return a copy of Options of a Session value.
// In case of data race.
func (s *Session) GetOptions() Options {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return *s.data.Options
}

// SetMaxAge sets the MaxAge of a session.
func (s *Session) SetMaxAge(seconds int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data.Options.MaxAge = time.Duration(seconds)
	// Set expiresTimestamp for deleting expired session.
	// Users don't need to care expiresTimestamp field of a session.
	s.data.Expiry = time.Now().Add(time.Duration(seconds) * time.Second).Unix()
}

func (s *Session) SetCookiePath(path string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data.Options.Path = path
}

func (s *Session) SetCookieDomain(domain string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data.Options.Domain = domain
}

func (s *Session) SetCookieSecure(secure bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data.Options.Secure = secure
}

func (s *Session) SetCookieHttpOnly(isHttpOnly bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data.Options.HttpOnly = isHttpOnly
}

func (s *Session) SetCookieSameSite(sameSite http.SameSite) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data.Options.SameSite = sameSite
}

func (s *Session) GetMaxAge() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return int(s.data.Options.MaxAge)
}

func (s *Session) GetCookiePath() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.data.Options.Path
}

func (s *Session) GetCookieDomain() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.data.Options.Domain
}

func (s *Session) GetCookieSecure() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.data.Options.Secure
}

func (s *Session) GetCookieHttpOnly() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.data.Options.HttpOnly
}

func (s *Session) GetCookieSameSite() http.SameSite {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.data.Options.SameSite
}
