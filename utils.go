package sessions

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"math/big"
	"net/http"
	"time"
)

// generateRandomID adopted from https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb
func generateRandomID(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz=!@#$%^&*()-_+"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", fmt.Errorf("failed to generate session id: %v", err)
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

// deepCopyMap performs a deep copy of the given map m.
//
// learn more: https://davidzhu.xyz/post/golang/basics/014-gob-json-encoding/#27-gobregister-method
func deepCopyMap(m map[string]interface{}) (map[string]interface{}, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	dec := gob.NewDecoder(buf)
	if err := enc.Encode(m); err != nil {
		return nil, fmt.Errorf("failed to copy map: %v", err)
	}
	result := make(map[string]interface{})
	if err := dec.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to copy map: %v", err)
	}
	return result, nil
}

// validateCookieOptions validates cookie options and returns an error if invalid
func validateCookieOptions(options *Options) error {
	if options.SameSite == http.SameSiteNoneMode && !options.Secure {
		return fmt.Errorf("cookies with SameSite=None must be Secure")
	}
	return nil
}

// defaultOptions returns the default configuration
func defaultOptions() *Options {
	return &Options{
		Path:     "/",
		MaxAge:   60 * time.Second,
		Secure:   true,                    // Changed default to true for better security
		HttpOnly: true,                    // Changed default to true for better security
		SameSite: http.SameSiteStrictMode, // Changed default to Strict for better security
	}
}

// Option functions for customizing MemoryStore

// WithOptions sets the cookie options for the store
func WithOptions(options *Options) func(store *MemoryStore) {
	return func(store *MemoryStore) {
		if options == nil {
			return
		}

		if err := options.Validate(); err != nil {
			panic(err)
		}

		store.options = options
	}
}

// WithSessionIDLength sets the session ID length
func WithSessionIDLength(length int) func(store *MemoryStore) {
	return func(store *MemoryStore) {
		if length <= 0 {
			panic("session ID length must be greater than 0")
		}
		store.idLength = length
	}
}

// WithGCInterval sets the garbage collection interval
func WithGCInterval(interval time.Duration) func(store *MemoryStore) {
	return func(store *MemoryStore) {
		if interval <= 0 {
			panic("GC interval must be greater than 0")
		}
		store.gcInterval = interval
	}
}
