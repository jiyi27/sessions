package sessions

import (
	"net/http"
	"time"
)

// NewCookie returns a http.Cookie with the options set.
// For Internet Explorer Browser compatibility reason, it also sets
// the Expires field calculated based on the MaxAge value.
// https://github.com/golang/go/issues/52989#issuecomment-1131176565
func NewCookie(name, value string, options *Options) *http.Cookie {
	cookie := newCookieFromOptions(name, value, options)
	if options.MaxAge > 0 {
		// This is a type conversion: convert 'int' to 'time.Duration' type which is 'int64'
		d := options.MaxAge * time.Second
		cookie.Expires = time.Now().Add(d)
	} else if options.MaxAge < 0 {
		// Set it to the past to expire now.
		cookie.Expires = time.Unix(1, 0)
	}
	return cookie
}

// newCookieFromOptions returns a http.Cookie with the options set.
func newCookieFromOptions(name, value string, options *Options) *http.Cookie {
	return &http.Cookie{
		Name:   name,
		Value:  value,
		Path:   options.Path,
		Domain: options.Domain,
		// Max-Age is relative to the time of setting, Expiration = Tsetting + Max-Age
		// So don't need to switch time zone
		// https://stackoverflow.com/a/35729939/16317008
		MaxAge:   int(options.MaxAge),
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
		SameSite: options.SameSite,
	}
}
