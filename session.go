package sessions

import "net/http"

// NewSession is called by session stores to create a new session instance.
func NewSession(name, id string, store Store) *Session {
	return &Session{
		name:    name,
		id:      id,
		Values:  make(map[interface{}]interface{}),
		IsNew:   true,
		Options: new(Options),
		store:   store,
	}
}

// Session stores the values and optional configuration for a session.
type Session struct {
	name string
	id   string
	// Values contain the user-data for the session.
	// Maps are reference types, so they are always passed by reference.
	// So don't need to save as a pointer here.
	Values  map[interface{}]interface{}
	Options *Options
	IsNew   bool
	store   Store
}

// Save is a convenience method to save this session. It is the same as calling
// store.Save(request, response, session). You should call Save before writing to
// the response or returning from the handler.
func (s *Session) Save(r *http.Request, w http.ResponseWriter) error {
	return s.store.Save(r, w, s)
}

// Name returns the name used to register the session.
func (s *Session) Name() string {
	return s.name
}

// Store returns the session store used to register the session.
func (s *Session) Store() Store {
	return s.store
}

type sessionInfo struct {
	session          *Session
	expiresTimestamp int64
}
