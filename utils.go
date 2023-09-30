package sessions

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"math/big"
)

// GenerateRandomString adopted from https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb
func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz=#"
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

// DeepCopyMap performs a deep copy of the given map m.
//
// learn more: https://davidzhu.xyz/post/golang/basics/014-gob-json-encoding/#27-gobregister-method
func DeepCopyMap(m map[interface{}]interface{}) (map[interface{}]interface{}, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	dec := gob.NewDecoder(buf)
	if err := enc.Encode(m); err != nil {
		return nil, fmt.Errorf("failed to copy map: %v", err)
	}
	result := make(map[interface{}]interface{})
	if err := dec.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to copy map: %v", err)
	}
	return result, nil
}
