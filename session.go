package sessions

// NewSession is called by session stores to create a new session instance.
func NewSession(name string, id string, opts *Options) *Session {
	return &Session{
		name:    name,
		id:      id,
		Values:  make(map[interface{}]interface{}),
		Options: opts,
		IsNew:   true,
	}
}

// Session stores the values and optional configuration for a session.
type Session struct {
	name string
	id   string
	// Values contains the user-data for the session.
	Values  map[interface{}]interface{}
	Options *Options
	IsNew   bool
}

type sessionInfo struct {
	session          *Session
	expiresTimestamp int64
}
