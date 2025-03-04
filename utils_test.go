// https://www.digitalocean.com/community/tutorials/how-to-write-unit-tests-in-go-using-go-test-and-the-testing-package
// Test file must always end with _test.go
// Adopted from: https://gist.github.com/soroushjp/0ec92102641ddfc3ad5515ca76405f4d
package sessions

import (
	"encoding/gob"
	"reflect"
	"testing"
)

// test function signature: func TestXxxx(t *testing.T).
func TestDeepCopyMap(t *testing.T) {
	testCases := []struct {
		// original and expectedOriginal are the same value in each test case.
		original         map[string]interface{}
		transformer      func(m map[string]interface{}) map[string]interface{}
		expectedCopy     map[string]interface{}
		expectedOriginal map[string]interface{}
	}{
		// reassignment of entire map, should be okay even without deep-copy.
		{
			original: nil,
			transformer: func(m map[string]interface{}) map[string]interface{} {
				return map[string]interface{}{}
			},
			expectedCopy:     map[string]interface{}{},
			expectedOriginal: nil,
		},
		// mutation of map
		{
			original: map[string]interface{}{
				"id":  "0007",
				"age": 3,
			},
			transformer: func(m map[string]interface{}) map[string]interface{} {
				m["id"] = "0006"
				return m
			},
			expectedCopy: map[string]interface{}{
				"id":  "0006",
				"age": 3,
			},
			expectedOriginal: map[string]interface{}{
				"id":  "0007",
				"age": 3,
			},
		},
		// mutation of nested maps
		{
			original: map[string]interface{}{
				"id": "0007",
				"cats": map[string]int{
					"kitten": 2,
					"milo":   1,
				},
			},
			transformer: func(m map[string]interface{}) map[string]interface{} {
				m["cats"].(map[string]int)["kitten"] = 3
				return m
			},
			expectedCopy: map[string]interface{}{
				"id": "0007",
				"cats": map[string]int{
					"kitten": 3,
					"milo":   1,
				},
			},
			expectedOriginal: map[string]interface{}{
				"id": "0007",
				"cats": map[string]int{
					"kitten": 2,
					"milo":   1,
				},
			},
		},
		// mutation of nested slices
		{
			original: map[string]interface{}{
				"cats": []string{"Coco", "Bella"},
			},
			transformer: func(m map[string]interface{}) map[string]interface{} {
				m["cats"].([]string)[0] = "Luna"
				return m
			},
			expectedCopy: map[string]interface{}{
				"cats": []string{"Luna", "Bella"},
			},
			expectedOriginal: map[string]interface{}{
				"cats": []string{"Coco", "Bella"},
			},
		},
	}
	gob.Register(map[string]int{})
	for i, tc := range testCases {
		result, err := deepCopyMap(tc.original)
		if err != nil {
			t.Fatalf("error happens: %v, in test case: %d", err, i)
		}
		tc.transformer(result)
		// reflect.DeepEqual(): https://stackoverflow.com/a/18211675/16317008
		// https://www.reddit.com/r/golang/comments/t3cbsv/comment/hysi819/?utm_source=share&utm_medium=web2x&context=3
		eq := reflect.DeepEqual(tc.transformer(result), tc.expectedCopy)
		if !eq {
			t.Errorf("copy was not mutated. test case: %d", i)
		}
		eq = reflect.DeepEqual(tc.original, tc.expectedOriginal)
		if !eq {
			t.Errorf("original was mutated. test case: %d", i)
		}
	}
}
