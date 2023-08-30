package sessions

// NewSession is called by session stores to create a new session instance.
func NewSession(name, id string) *Session {
	return &Session{
		name:    name,
		id:      id,
		Values:  make(map[interface{}]interface{}),
		IsNew:   true,
		Options: new(Options),
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
}

type sessionInfo struct {
	session          *Session
	expiresTimestamp int64
}
