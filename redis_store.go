package sessions

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore represents a session store backed by Redis.
type RedisStore struct {
	client     *redis.Client
	options    *Options
	idLength   int
	serializer *Serializer
}

// NewRedisStore creates a new RedisStore with the given Redis client and options.
func NewRedisStore(client *redis.Client, options ...func(*RedisStore)) *RedisStore {
	s := &RedisStore{
		client: client,
		options: &Options{
			Path:     "/",
			MaxAge:   86400,
			SameSite: http.SameSiteDefaultMode,
		},
		idLength:   16,
		serializer: &Serializer{},
	}

	for _, op := range options {
		op(s)
	}

	return s
}

//// WithIDLength sets the session ID length.
//func WithIDLength(l int) func(*RedisStore) {
//	return func(s *RedisStore) {
//		s.idLength = l
//	}
//}
//
//// WithMaxAge sets the maximum age of the session.
//func WithMaxAge(maxAge int) func(*RedisStore) {
//	return func(s *RedisStore) {
//		s.options.MaxAge = maxAge
//	}
//}

// generateID generates a unique session ID.
func (s *RedisStore) generateID() (string, error) {
	for {
		id, err := GenerateRandomString(s.idLength)
		if err != nil {
			return "", err
		}
		exists, err := s.client.Exists(context.Background(), id).Result()
		if err != nil {
			return "", err
		}
		if exists == 0 {
			return id, nil
		}
	}
}

// Get retrieves a session by name from the Redis store or creates a new one.
func (s *RedisStore) Get(r *http.Request, name string) (*Session, error) {
	if !isCookieNameValid(name) {
		return nil, fmt.Errorf("sessions: invalid character in cookie name: %s", name)
	}

	cookie, err := r.Cookie(name)
	if err != nil {
		return s.New(name)
	}

	sessionID := cookie.Value
	data, err := s.client.Get(context.Background(), sessionID).Result()
	if err != nil && errors.Is(err, redis.Nil) {
		return s.New(name)
	} else if err != nil {
		return nil, err
	}

	session := &Session{}
	err = s.serializer.Deserialize([]byte(data), session)
	if err != nil {
		return nil, err
	}
	session.data.IsNew = false
	return session, nil
}

// New creates a new session and saves it in the Redis store.
func (s *RedisStore) New(name string) (*Session, error) {
	id, err := s.generateID()
	if err != nil {
		return nil, err
	}

	session := NewSession(name, id, *s.options)
	err = s.Save(session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// Save persists the session in the Redis store.
func (s *RedisStore) Save(session *Session) error {
	data, err := s.serializer.Serialize(session)
	if err != nil {
		return err
	}
	expiration := time.Duration(session.data.Options.MaxAge) * time.Second
	return s.client.Set(context.Background(), session.data.ID, data, expiration).Err()
}

// Delete removes the session from the Redis store.
func (s *RedisStore) Delete(session *Session) error {
	return s.client.Del(context.Background(), session.data.ID).Err()
}
