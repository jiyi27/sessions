package sessions

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	// Creating and initializing a map
	m_a_p := map[int]bool{
		90: true,
		91: false,
	}
	// Using blank identifier
	a, ok1 := m_a_p[95]
	fmt.Println("\nKey present or not:", ok1)
	fmt.Println(a)
}
