package frames_test

import (
	"fmt"
	"testing"

	"github.com/nanobus/iota/go/internal/frames"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetup(t *testing.T) {
	s := frames.Setup{
		MajorVersion: 0,
		MinorVersion: 2,
		Data:         []byte("testing1234"),
	}

	size := s.Size()
	buf := make([]byte, size)
	s.Encode(buf)

	fmt.Println(buf[frames.FrameHeaderLen:])

	var s2 frames.Setup
	f := frames.ParseFrameHeader(buf)
	require.NoError(t, s2.Decode(&f, buf[frames.FrameHeaderLen:]))

	assert.Equal(t, s2, s)
}
