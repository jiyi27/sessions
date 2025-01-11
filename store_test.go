package sessions

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// BenchmarkMemoryStore_Concurrent tests the concurrent performance
func BenchmarkMemoryStore_ConcurrentAccess(b *testing.B) {
	store := NewMemoryStore(
		WithMaxAge(3600),
		WithGCInterval(time.Second),
	)

	// 先创建一些共享的 sessions
	sessions := make([]*Session, 100)
	for i := 0; i < 100; i++ {
		session, _ := store.New(fmt.Sprintf("test_session_%d", i))
		sessions[i] = session
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		// 每个 goroutine 创建自己的 request
		req := httptest.NewRequest("GET", "http://example.com", nil)

		for pb.Next() {
			// 随机选择一个已存在的 session
			idx := rand.Intn(len(sessions))
			session := sessions[idx]

			// 设置 cookie
			cookie := &http.Cookie{
				Name:  session.name,
				Value: session.id,
			}
			req.Header.Set("Cookie", cookie.String())

			// 并发读取同一个 session
			_, err := store.Get(req, session.name)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestStore(t *testing.T) {
	var req *http.Request
	var rsp *httptest.ResponseRecorder
	var hdr http.Header
	var err error
	var ok bool
	var cookies []string
	var session *Session
	store := NewMemoryStore()

	// Round 1 ----------------------------------------------------------------

	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	rsp = httptest.NewRecorder()
	// Get a session.
	if session, err = store.Get(req, "session-key"); err != nil {
		t.Fatalf("Error getting session: %v", err)
	}
	session.SetValue("name", "Coco")
	session.SetValue("age", 18)
	session.SetMaxAge(3)
	session.Save(rsp)
	hdr = rsp.Header()
	cookies, ok = hdr["Set-Cookie"]
	if !ok || len(cookies) != 1 {
		t.Fatal("No cookies. Header:", hdr)
	}

	// Round 2 ----------------------------------------------------------------

	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	// Simulate client send cookie that we've saved into the response in last round.
	req.Header.Add("Cookie", cookies[0])
	if session, err = store.Get(req, "session-key"); err != nil {
		t.Fatalf("Error getting session: %v", err)
	}
	// Test if the session was saved successfully in last round.
	if session.IsNew() {
		t.Errorf("Expected session.isNew=false; got, session.isNew=%v", session.IsNew())
	}
	// Test if the gc deletes session incorrectly.
	time.Sleep(time.Second)
	if session, err = store.Get(req, "session-key"); err != nil {
		t.Fatalf("Error getting session: %v", err)
	}
	if session.IsNew() {
		t.Fatal("gc deletes session incorrectly")
	}
	_ = session.GetValueByKey("name")
	_ = session.GetValueByKey("age")
	// Modify value for next round test.
	session.SetValue("name", "Bella")
	session.SetCookieHttpOnly(true)
	session.Save(rsp)

	// Round 3 ----------------------------------------------------------------

	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	req.Header.Add("Cookie", cookies[0])
	if session, err = store.Get(req, "session-key"); err != nil {
		t.Fatalf("Error getting session: %v", err)
	}
	name := session.GetValueByKey("name")
	if name != "Bella" {
		t.Errorf("Expected name = Bella; Got name=%v", name)
	}
	if !session.GetCookieHttpOnly() {
		t.Errorf("Expected http only = true; httponly=%v", session.GetCookieHttpOnly())
	}
	// Test if sessions are removed correctly
	time.Sleep(5 * time.Second)
	if session, err = store.Get(req, "session-key"); err != nil {
		t.Fatalf("Error getting session: %v", err)
	}
	if !session.IsNew() {
		t.Errorf("Expected session.IsNew() = true; Got session.IsNew=%v", session.IsNew())
	}
}
