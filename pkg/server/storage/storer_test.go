package storage

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestType_String(t *testing.T) {
	tps := []Type{Unspecified, Memory, DataStore, Postgres}
	for _, tp := range tps {
		assert.NotEmpty(t, tp.String())
	}
}
