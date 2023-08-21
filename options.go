package sessions

// Options stores configuration for a session or session store.
//
// Fields are a subset of http.Cookie fields.
type Options struct {
	Path   string
	Domain string
	// MaxAge specified in seconds
	MaxAge   int
	Secure   bool
	HttpOnly bool
}
