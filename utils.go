package sessions

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb
func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz=#-+%@!~"
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
