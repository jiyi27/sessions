package sessions

import (
	"errors"
	"fmt"
	"net/http"
)

// Store interface defines the contract for session storage implementations
type Store interface {
	// Get retrieves an existing session or creates a new one
	// Returns error if the cookie name is invalid or if any other error occurs
	Get(r *http.Request, name string) (*Session, error)

	// New creates a new session with the given name
	// Returns error if session ID generation fails or if any other error occurs
	New(name string) (*Session, error)

	// Save persists the session data to the store
	// Returns error if serialization or storage operation fails
	Save(session *Session) error

	// Delete removes the session from the store
	// Returns error if the deletion operation fails
	Delete(session *Session) error
}

// baseStore implements common functionality for all stores
type baseStore struct {
	options  *Options // default cookie options value when creating a new session
	idLength int      // length of the session ID
}

// NewBaseStore creates a new baseStore with default options
func newBaseStore(opts *Options, idLen int) (*baseStore, error) {
	if opts == nil {
		opts = defaultOptions()
	}

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	if idLen <= 0 {
		return nil, errors.New("id length must be positive")
	}

	return &baseStore{
		options:  opts,
		idLength: idLen,
	}, nil
}
