package sessions

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func TestRedisStore_Get(t *testing.T) {
	client := setupRedisClient()
	store := NewRedisStore(client)

	sessionID := "test_session_id"
	sessionData := `{"ID":"test_session_id","Name":"test_session","Values":{}}`
	client.Set(context.Background(), sessionID, sessionData, 60*time.Second)

	req, _ := http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "test_session", Value: sessionID})

	session, err := store.Get(req, "test_session")
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, sessionID, session.data.ID)
	assert.False(t, session.data.IsNew)
}

func TestRedisStore_New(t *testing.T) {
	client := setupRedisClient()
	store := NewRedisStore(client)

	session, err := store.New("new_session")
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "new_session", session.data.Name)
	assert.True(t, session.data.IsNew)

	// 验证新会话是否已保存到 Redis
	data, err := client.Get(context.Background(), session.data.ID).Result()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}
