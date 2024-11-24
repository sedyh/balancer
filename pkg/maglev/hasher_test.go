package maglev

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventualConsistency(t *testing.T) {
	tests := 20
	passes := 200
	file := "file-0:part-0"
	first := []string{"backend-1", "backend-2", "backend-3"}
	second := []string{"backend-4", "backend-5", "backend-6"}
	destinations := make([]string, 0, passes)

	for i := 0; i < tests; i++ {
		hash := NewHasher(DefaultPrime)

		hash.AddBackends(first)
		destinations = destinations[:0]
		for i := 0; i < passes; i++ {
			destinations = append(destinations, hash.GetBackend(file))
		}
		for i := 1; i < len(destinations); i++ {
			assert.Equal(t, destinations[i], destinations[0])
		}

		hash.AddBackends(second)
		destinations = destinations[:0]
		for i := 0; i < passes; i++ {
			destinations = append(destinations, hash.GetBackend(file))
		}
		for i := 1; i < len(destinations); i++ {
			assert.Equal(t, destinations[i], destinations[0])
		}
	}
}
