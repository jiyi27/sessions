package main

import (
	"fmt"
	"log"
	"net/http"
	"sessions"
)

func DoNothing(_ http.ResponseWriter, _ *http.Request) {}

func handler(w http.ResponseWriter, r *http.Request) {
	// Get will return a cached session if exists, if not return a new one.
	// For simplicity, ignore error here
	session, _ := store.Get(r, "session-id")
	// If it's a new session, set some info into it.
	if session.IsNew() {
		session.InsertValue("name", "Coco")
		session.InsertValue("age", 18)
		// Set max age for session in seconds.
		session.SetMaxAge(30)
		// Save session into response to client.
		// You should save session after you make change on a session.
		session.Save(w)
		return
	}
	// If not a new session.
	name, _ := session.GetValueByKey("name")
	_, _ = fmt.Fprint(w, fmt.Sprintf("hello %v\n", name))
}

// You just need only one store instance on global.
var store *sessions.MemoryStore

func main() {
	// You need specify the id length of a session, don't make it too big.
	store = NewMemoryStore(WithExpiredSessionTracking())
	http.HandleFunc("/", handler)
	http.HandleFunc("/favicon.ico", DoNothing)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
