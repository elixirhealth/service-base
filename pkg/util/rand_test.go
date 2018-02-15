package util

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandBytes(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	lens := []int{1, 2, 4, 8, 16}

	for _, l := range lens {
		b := RandBytes(rng, l)
		assert.Len(t, b, l)
		assert.NotEqual(t, make([]byte, l), b)
	}
}
