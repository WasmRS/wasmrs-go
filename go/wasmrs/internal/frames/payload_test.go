package frames

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPayloadFragment(t *testing.T) {
	const maxFrameSize = 1024 + FrameHeaderLen + 3
	data := make([]byte, 1024*1024)
	metadata := make([]byte, 1024*1024)

	fill(data)
	fill(metadata)

	p := Payload{
		StreamID: 12345,
		Metadata: metadata,
		Data:     data,
		Complete: true,
		Next:     true,
	}

	frames := p.Fragment(maxFrameSize)
	require.NotNil(t, frames)
	const expectedFrames = 1024 + 1022
	require.Lenf(t, frames, expectedFrames,
		"expected %d fragments. got %d", expectedFrames, len(frames))

	var actualMD []byte
	var actualData []byte

	for i, f := range frames {
		pl := f.(*Payload)
		actualMD = append(actualMD, pl.Metadata...)
		actualData = append(actualData, pl.Data...)
		assert.Equal(t, i == expectedFrames-1, pl.Complete)
		assert.Equal(t, i == expectedFrames-1, pl.Next)
		if i != expectedFrames-1 {
			assert.Equal(t, pl.Size(), maxFrameSize)
		}
	}

	assert.Equal(t, metadata, actualMD)
	assert.Equal(t, data, actualData)
}

func fill(buf []byte) {
	var b byte
	for i := 0; i < len(buf); i++ {
		buf[i] = b
		b++
	}
}
