package sessions

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Learn more about benchmark:
// https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go

func BenchmarkSession(b *testing.B) {
	store := newCookieStore()
	// The body function will be run in each goroutine.
	b.RunParallel(func(pb *testing.PB) {
		var req *http.Request
		var rsp *httptest.ResponseRecorder
		var hdr http.Header
		var err error
		var ok bool
		var cookies []string
		var session *Session
		for pb.Next() {
			req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
			rsp = httptest.NewRecorder()
			if session, err = store.Get(req, "session-key"); err != nil {
				b.Fatalf("Error getting session: %v", err)
			}
			// Simulate user set information in session.
			session.InsertValue("name", "Coco")
			session.InsertValue("age", 18)
			session.SetMaxAge(1)
			session.Save(rsp)
			hdr = rsp.Header()
			cookies, ok = hdr["Set-Cookie"]
			if !ok || len(cookies) != 1 {
				b.Fatal("No cookies. Header:", hdr)
			}
			req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
			// Simulate client send cookie that we've saved into the response in last round.
			req.Header.Add("Cookie", cookies[0])
			if session, err = store.Get(req, "session-key"); err != nil {
				b.Fatalf("Error getting session: %v", err)
			}
			// Simulate user gets info from session.
			session.IsNew()
			_, _ = session.GetValueByKey("name")
			_, _ = session.GetValueByKey("age")
			// Simulate user changes the session.
			err = session.ModifyValueByKey("name", "Bella")
			session.SetCookieHttpOnly(true)
			session.SetMaxAge(5)
			session.Save(rsp)
		}
	})
}

func BenchmarkSessionWithDeepCopy(b *testing.B) {
	store := newMemoryStore()
	// The body function will be run in each goroutine.
	b.RunParallel(func(pb *testing.PB) {
		var req *http.Request
		var rsp *httptest.ResponseRecorder
		var hdr http.Header
		var err error
		var ok bool
		var cookies []string
		var session *MySession
		for pb.Next() {
			req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
			rsp = httptest.NewRecorder()
			if session, err = store.Get(req, "session-key"); err != nil {
				b.Fatalf("Error getting session: %v", err)
			}
			// Simulate user set information in session.
			session.Values["name"] = "Coco"
			session.Values["age"] = 18
			session.Options.MaxAge = 1
			_ = session.Save(req, rsp)
			hdr = rsp.Header()
			cookies, ok = hdr["Set-Cookie"]
			if !ok || len(cookies) != 1 {
				b.Fatal("No cookies. Header:", hdr)
			}
			req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
			// Simulate client send cookie that we've saved into the response in last round.
			req.Header.Add("Cookie", cookies[0])
			if session, err = store.Get(req, "session-key"); err != nil {
				b.Fatalf("Error getting session: %v", err)
			}
			// Simulate user gets info from session.
			_ = session.IsNew
			_, _ = session.Values["name"]
			_, _ = session.Values["age"]
			// Simulate user changes the session.
			session.Values["name"] = "Bella"
			session.Options.HttpOnly = true
			session.Options.MaxAge = 5
			_ = session.Save(req, rsp)
		}
	})
}
