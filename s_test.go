package sessions

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSession(t *testing.T) {
	var req *http.Request
	var rsp *httptest.ResponseRecorder
	var hdr http.Header
	var err error
	var ok bool
	var cookies []string
	var session *Session

	store := newCookieStore()

	// Round 1 ----------------------------------------------------------------

	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	rsp = httptest.NewRecorder()
	// Get a session.
	if session, err = store.Get(req, "session-key"); err != nil {
		t.Fatalf("Error getting session: %v", err)
	}
	session.InsertValue("name", "Coco")
	session.InsertValue("age", 18)
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
	_, err1 := session.GetValueByKey("name")
	_, err2 := session.GetValueByKey("age")
	if err1 != nil || err2 != nil {
		t.Fatal("Error the vlaues has not been saved successfully")
	}
	// Modify value for next round test.
	err = session.ModifyValueByKey("name", "Bella")
	session.SetCookieHttpOnly(true)
	session.Save(rsp)

	// Round 3 ----------------------------------------------------------------

	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	req.Header.Add("Cookie", cookies[0])
	if session, err = store.Get(req, "session-key"); err != nil {
		t.Fatalf("Error getting session: %v", err)
	}
	name, err := session.GetValueByKey("name")
	if name != "Bella" {
		t.Errorf("Expected name = Bella; Got name=%v", name)
	}
	if !session.GetCookieHttpOnly() {
		t.Errorf("Expected http only = true; httponly=%v", session.GetCookieHttpOnly())
	}
	// Test if sessions are removed correctly
	time.Sleep(3 * time.Second)
	if session, err = store.Get(req, "session-key"); err != nil {
		t.Fatalf("Error getting session: %v", err)
	}
	if !session.IsNew() {
		t.Errorf("Expected session.IsNew() = true; Got session.IsNew=%v", session.IsNew())
	}
}
