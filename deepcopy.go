package sessions

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

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
