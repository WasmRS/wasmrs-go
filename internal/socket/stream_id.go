package socket

import (
	"sync/atomic"
)

const (
	_maskStreamID uint64 = 0x7FFFFFFF
	_halfSeed     uint64 = 0x40000000
)

// StreamID can be used to generate stream ids.
type StreamID interface {
	// Next returns next stream id.
	Next() (id uint32, firstLoop bool)
}

type ServerStreamIDs uint64

func (p *ServerStreamIDs) Next() (uint32, bool) {
	// 2,4,6,8...
	seed := atomic.AddUint64((*uint64)(p), 1)
	v := 2 * seed
	if v != 0 {
		return uint32(_maskStreamID & v), seed <= _halfSeed
	}
	return p.Next()
}

type ClientStreamIDs uint64

func (p *ClientStreamIDs) Next() (uint32, bool) {
	// 1,3,5,7
	seed := atomic.AddUint64((*uint64)(p), 1)
	v := 2*(seed-1) + 1
	if v != 0 {
		return uint32(_maskStreamID & v), seed <= _halfSeed
	}
	return p.Next()
}
