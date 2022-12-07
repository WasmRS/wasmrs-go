package rsocket

import (
	"bufio"
	"errors"
	"io"

	"github.com/nanobus/iota/go/internal/frames"
	"github.com/nanobus/iota/go/internal/u24"
)

const (
	lengthFieldSize = 3
	minBuffSize     = 8 * 1024
	maxBuffSize     = 16*1024*1024 + lengthFieldSize
)

// ErrIncompleteHeader is error of incomplete header.
var ErrIncompleteHeader = errors.New("incomplete frame header")

// from core, errors.go
var ErrInvalidFrameLength = errors.New("rsocket: invalid frame length")

// LengthBasedFrameDecoder defines a decoder for decoding frames which have a header of length.
type LengthBasedFrameDecoder bufio.Scanner

// Read reads next raw frame in bytes.
func (p *LengthBasedFrameDecoder) Read() (raw []byte, err error) {
	scanner := (*bufio.Scanner)(p)
	if !scanner.Scan() {
		err = scanner.Err()
		if err == nil || isClosedErr(err) {
			err = io.EOF
		}
		return
	}
	raw = scanner.Bytes()[lengthFieldSize:]
	if len(raw) < frames.FrameHeaderLen {
		err = ErrIncompleteHeader
	}
	return
}

func doSplit(data []byte, eof bool) (advance int, token []byte, err error) {
	if eof {
		return
	}
	if len(data) < lengthFieldSize {
		return
	}
	frameLength := u24.ReadUint24ToInt(data)
	if frameLength < 1 {
		err = ErrInvalidFrameLength
		return
	}
	frameSize := frameLength + lengthFieldSize
	if frameSize <= len(data) {
		return frameSize, data[:frameSize], nil
	}
	return
}

// NewLengthBasedFrameDecoder creates a new frame decoder.
func NewLengthBasedFrameDecoder(r io.Reader) *LengthBasedFrameDecoder {
	scanner := bufio.NewScanner(r)
	scanner.Split(doSplit)
	buf := make([]byte, 0, minBuffSize)
	scanner.Buffer(buf, maxBuffSize)
	return (*LengthBasedFrameDecoder)(scanner)
}
