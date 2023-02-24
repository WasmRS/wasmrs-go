package rsocket

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/nanobus/iota/go/internal/frames"
	"github.com/nanobus/iota/go/internal/u24"
)

// TCPConn is RSocket connection for TCP transport.
type TCPConn struct {
	conn    net.Conn
	writer  *bufio.Writer
	decoder *LengthBasedFrameDecoder
	// counter *core.TrafficCounter
}

func (p TCPConn) Addr() string {
	addr := p.conn.RemoteAddr()
	return addr.String()
}

// // SetCounter bind a counter which can count r/w bytes.
// func (p *TCPConn) SetCounter(c *core.TrafficCounter) {
// 	p.counter = c
// }

// SetDeadline set deadline for current connection.
// After this deadline, connection will be closed.
func (p *TCPConn) SetDeadline(deadline time.Time) error {
	return p.conn.SetReadDeadline(deadline)
}

// Read reads next frame from Conn.
func (p *TCPConn) Read() (f frames.Frame, err error) {
	raw, err := p.decoder.Read()
	if err == io.EOF {
		return
	}
	if err != nil {
		err = fmt.Errorf("read frame failed: %w", err)
		return
	}

	header := frames.ParseFrameHeader(raw)
	buffer := raw[frames.FrameHeaderLen:]

	switch header.Type() {
	case frames.FrameTypeSetup:
		var setup frames.Setup
		if err := setup.Decode(&header, buffer); err != nil {
			return nil, err
		}
		return &setup, nil

	case frames.FrameTypeRequestResponse, frames.FrameTypeRequestFNF,
		frames.FrameTypeRequestStream, frames.FrameTypeRequestChannel:
		var rr frames.RequestPayload
		if err := rr.Decode(&header, buffer); err != nil {
			return nil, err
		}
		return &rr, nil

	case frames.FrameTypeRequestN:
		var rn frames.RequestN
		if err := rn.Decode(&header, buffer); err != nil {
			return nil, err
		}
		return &rn, nil

	case frames.FrameTypeCancel:
		var cancel frames.Cancel
		if err := cancel.Decode(&header, buffer); err != nil {
			return nil, err
		}
		return &cancel, nil

	case frames.FrameTypePayload:
		var p frames.Payload
		if err := p.Decode(&header, buffer); err != nil {
			return nil, err
		}
		return &p, nil

	case frames.FrameTypeError:
		var e frames.Error
		if err := e.Decode(&header, buffer); err != nil {
			return nil, err
		}
		return &e, nil

	default:
		return nil, fmt.Errorf("unknown frame %d", header.Type())
	}
}

// Write writes a frame.
func (p *TCPConn) Write(frame frames.Frame) (err error) {
	size := frame.Size()
	// if p.counter != nil && frame.Header().Resumable() {
	// 	p.counter.IncWriteBytes(size)
	// }
	_, err = u24.MustNewUint24(size).WriteTo(p.writer)
	if err != nil {
		err = fmt.Errorf("write frame failed: %w", err)
		return
	}

	// TODO: remove alloc
	buf := make([]byte, size)
	frame.Encode(buf)
	_, err = p.writer.Write(buf)
	// _, err = frame.WriteTo(p.writer)
	if err != nil {
		err = fmt.Errorf("write frame failed: %w", err)
		return
	}

	return
}

// Flush flush data.
func (p *TCPConn) Flush() (err error) {
	err = p.writer.Flush()
	if err != nil {
		err = fmt.Errorf("flush failed: %w", err)
	}
	return
}

// Close closes current connection.
func (p *TCPConn) Close() error {
	return p.conn.Close()
}

// NewTCPConn creates a new TCP RSocket connection.
func NewTCPConn(conn net.Conn) *TCPConn {
	return &TCPConn{
		conn:    conn,
		writer:  bufio.NewWriterSize(conn, 8192),
		decoder: NewLengthBasedFrameDecoder(conn),
	}
}
