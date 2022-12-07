package rsocket

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/nanobus/iota/go/internal/frames"
)

// ListenerFactory is factory which generate new listeners.
type ListenerFactory func(context.Context) (net.Listener, error)

type HandlerFactory func(context.Context) (DuplexHandler, error)

// Conn is connection for RSocket.
type Conn interface {
	io.Closer
	// SetDeadline set deadline for current connection.
	// After this deadline, connection will be closed.
	SetDeadline(deadline time.Time) error
	// SetCounter bind a counter which can count r/w bytes.
	// SetCounter(c *core.TrafficCounter)
	// Read reads next frame from Conn.
	Read() (frames.Frame, error)
	// Write writes a frame to Conn.
	Write(frames.Frame) error
	// Flush flushes the data.
	Flush() error
}

type AddrConn interface {
	Conn
	Addr() string
}
