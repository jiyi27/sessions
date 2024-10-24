package sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client    *redis.Client
	options   *Options
	idLength  int
	keyPrefix string
}

func NewRedisStore(redisClient *redis.Client) *RedisStore {
	return &RedisStore{
		client: redisClient,
		options: &Options{
			Path:     "/",
			MaxAge:   60,
			SameSite: http.SameSiteDefaultMode,
		},
		idLength:  16,
		keyPrefix: "session:",
	}
}

// Get 获取或创建会话
func (s *RedisStore) Get(r *http.Request, name string) (*Session, error) {
	if !isCookieNameValid(name) {
		return nil, fmt.Errorf("sessions: invalid character in cookie name: %s", name)
	}

	if c, err := r.Cookie(name); err == nil {
		// 从Redis中获取会话
		ctx := context.Background()
		data, err := s.client.Get(ctx, s.keyPrefix+c.Value).Bytes()
		if err == nil {
			session := NewSession(name, c.Value, *s.options)
			if err := json.Unmarshal(data, &session.values); err != nil {
				return nil, err
			}
			session.SetIsNew(false)
			return session, nil
		}
	}

	// 创建新会话
	return s.New(name)
}

// New 创建新会话
func (s *RedisStore) New(name string) (*Session, error) {
	id, err := s.generateID()
	if err != nil {
		return nil, err
	}
	session := NewSession(name, id, *s.options)
	
	// 保存到Redis
	ctx := context.Background()
	data, err := json.Marshal(session.values)
	if err != nil {
		return nil, err
	}
	
	err = s.client.Set(ctx, s.keyPrefix+session.id, data, time.Duration(s.options.MaxAge)*time.Second).Err()
	if err != nil {
		return nil, err
	}
	
	return session, nil
}

// Save 保存会话数据
func (s *RedisStore) Save(r *http.Request, w http.ResponseWriter, session *Session) error {
	ctx := context.Background()
	data, err := json.Marshal(session.values)
	if err != nil {
		return err
	}

	err = s.client.Set(ctx, s.keyPrefix+session.id, data, time.Duration(s.options.MaxAge)*time.Second).Err()
	if err != nil {
		return err
	}

	// 设置cookie
	session.mu.RLock()
	opts := *session.options
	session.mu.RUnlock()
	http.SetCookie(w, NewCookie(session.name, session.id, &opts))
	return nil
}

func (s *RedisStore) generateID() (string, error) {
	for {
		if id, err := GenerateRandomString(s.idLength); err != nil {
			return "", err
		} else {
			ctx := context.Background()
			exists, err := s.client.Exists(ctx, s.keyPrefix+id).Result()
			if err != nil {
				return "", err
			}
			if exists == 0 {
				return id, nil
			}
		}
	}
}
