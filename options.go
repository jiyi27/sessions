package sessions

import "net/http"

// Options stores configuration for a session or session store.
//
// Fields are a subset of http.Cookie fields.
type Options struct {
	Path   string
	Domain string
	// MaxAge is specified in seconds
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
}
