package socket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientStreamIDs_Next(t *testing.T) {
	var ids ClientStreamIDs
	id, firstLap := ids.Next()
	assert.Equal(t, uint32(1), id)
	assert.True(t, firstLap)
	id, firstLap = ids.Next()
	assert.Equal(t, uint32(3), id)
	assert.True(t, firstLap)
	id, firstLap = ids.Next()
	assert.Equal(t, uint32(5), id)
	assert.True(t, firstLap)
}

func TestServerStreamIDs_Next(t *testing.T) {
	var ids ServerStreamIDs
	id, firstLap := ids.Next()
	assert.Equal(t, uint32(2), id)
	assert.True(t, firstLap)
	id, firstLap = ids.Next()
	assert.Equal(t, uint32(4), id)
	assert.True(t, firstLap)
	id, firstLap = ids.Next()
	assert.Equal(t, uint32(6), id)
	assert.True(t, firstLap)
}

func BenchmarkServerStreamIDs_Next(b *testing.B) {
	var ids ServerStreamIDs
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ids.Next()
		}
	})
}

func BenchmarkClientStreamIDs_Next(b *testing.B) {
	var ids ClientStreamIDs
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ids.Next()
		}
	})
}
