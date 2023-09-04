package sessions

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSessionValues(t *testing.T) {
	var req *http.Request
	var rsp *httptest.ResponseRecorder
	var hdr http.Header
	var err error
	var ok bool
	var cookies []string
	var session *Session

	store := newMemoryStore()

	// Round 1 ----------------------------------------------------------------

	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	rsp = httptest.NewRecorder()
	// Get a session.
	if session, err = store.Get(req, "session-key"); err != nil {
		t.Fatalf("Error getting session: %v", err)
	}
	session.Values["name"] = "Coco"
	session.Values["age"] = 18
	session.Options.MaxAge = 3
	// Save.
	if err = session.Save(req, rsp); err != nil {
		t.Fatalf("Error saving session: %v", err)
	}
	hdr = rsp.Header()
	cookies, ok = hdr["Set-Cookie"]
	if !ok || len(cookies) != 1 {
		t.Fatal("No cookies. Header:", hdr)
	}

	// Round 2 ----------------------------------------------------------------

	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	// simulate client can send cookie that we've saved into its response in last round
	req.Header.Add("Cookie", cookies[0])
	// Get a session.
	if session, err = store.Get(req, "session-key"); err != nil {
		t.Fatalf("Error getting session: %v", err)
	}
	if session.IsNew {
		t.Errorf("Expected session.IsNew = false; Got session.IsNew=%v", session.IsNew)
	}
	// Test if the Values has been saved in last round
	if session.Values["name"] != "Coco" || session.Values["age"] != 18 {
		t.Errorf("Expected name=Coco,age=13; Got %v", session.Values)
	}
	// Test if there is a deep copy for session
	session.Values["name"] = "Bella"
	c, err := req.Cookie("session-key")
	if err != nil {
		t.Error("failed to get cookie")
	}
	memoryMutex.RLock()
	if session.Values["name"] == store.sessions[c.Value].session.Values["name"] {
		t.Errorf("No deep copy; Expected name=Coco; Got %v", store.sessions[c.Value].session.Values["name"])
	}
	memoryMutex.RUnlock()

	// Test if sessions are removed correctly
	time.Sleep(5 * time.Second)
	if session, err = store.Get(req, "session-key"); err != nil {
		t.Fatalf("Error getting session: %v", err)
	}
	if !session.IsNew {
		t.Errorf("Expected session.IsNew = true; Got session.IsNew=%v", session.IsNew)
	}
}
