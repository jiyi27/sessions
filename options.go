package sessions

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Options stores configuration for a session or session store.
//
// Fields are a subset of http.Cookie fields.
type Options struct {
	Path     string
	Domain   string
	MaxAge   time.Duration
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
}

// Validate checks if options are valid
func (o *Options) Validate() error {
	if o.Path == "" {
		return errors.New("path cannot be empty")
	}
	if o.MaxAge < 0 {
		return errors.New("max age cannot be negative")
	}
	if o.SameSite == http.SameSiteNoneMode && !o.Secure {
		return fmt.Errorf("cookies with SameSite=None must be Secure")
	}
	return nil
}
