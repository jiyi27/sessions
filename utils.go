package sessions

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
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

// DeepCopyMap performs a deep copy of the given map m.
// You should register for that type if you have nested map
// e.g., if you have map[string]int{} as a value of a key,
// then you need register: gob.Register(map[string]int{}) before call DeepCopyMap.
func DeepCopyMap(m map[interface{}]interface{}) (map[interface{}]interface{}, error) {
	gob.Register(map[interface{}]interface{}{})
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
