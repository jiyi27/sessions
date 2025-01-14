package sessions

import "net/http"

// Options stores configuration for a session or session store.
//
// Fields are a subset of http.Cookie fields.
type Options struct {
	Path     string
	Domain   string
	MaxAge   int // seconds
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
}
