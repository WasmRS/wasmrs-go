package rsocket

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"

	"github.com/nanobus/iota/go/internal/buffer"
	"github.com/nanobus/iota/go/internal/frames"
	"github.com/nanobus/iota/go/invoke"
)

var (
	errTransportClosed = errors.New("transport closed")
	errNoHandler       = errors.New("you must register a handler")
)

const (
	// DefaultKeepaliveInterval is default keepalive interval duration.
	DefaultKeepaliveInterval = 20 * time.Second
	// DefaultKeepaliveMaxLifetime is default keepalive max lifetime.
	DefaultKeepaliveMaxLifetime = 90 * time.Second
)

// FrameHandler is an alias of frame handler.
type FrameHandler = func(frame frames.Frame) (err error)

type DuplexHandler interface {
	SetFrameSender(handler FrameHandler)
	HandleFrame(frame frames.Frame) (err error)
}

// ServerTransportAcceptor is an alias of server transport handler.
type ServerTransportAcceptor = func(ctx context.Context, caller invoke.Caller, onClose func(*Transport))

// ServerTransport is server-side RSocket transport.
type ServerTransport interface {
	io.Closer
	// Accept register incoming connection handler.
	Accept(acceptor ServerTransportAcceptor)
	// Listen listens on the network address addr and handles requests on incoming connections.
	// You can specify notifier chan, it'll be sent true/false when server listening success/failed.
	Listen(ctx context.Context, notifier chan<- bool) error
}

// Transport is RSocket transport which is used to carry RSocket frames.
type Transport struct {
	conn        Conn
	maxLifetime time.Duration
	once        sync.Once
	handler     DuplexHandler
	isServer    bool
	ready       chan struct{}
}

// NewTransport creates new transport.
func NewTransport(c Conn, handler DuplexHandler, isServer bool) *Transport {
	t := Transport{
		conn:        c,
		maxLifetime: DefaultKeepaliveMaxLifetime,
		handler:     handler,
		isServer:    isServer,
		ready:       make(chan struct{}),
	}
	handler.SetFrameSender(t.handler.HandleFrame)
	return &t
}

// IsNoHandlerError returns true if input error means no handler registered.
func IsNoHandlerError(err error) bool {
	return err == errNoHandler
}

func (p *Transport) Addr() (string, bool) {
	ac, ok := p.conn.(AddrConn)
	if ok {
		return ac.Addr(), true
	}
	return "", false
}

// Connection returns current connection.
func (p *Transport) Connection() Conn {
	return p.conn
}

// SetLifetime set max lifetime for current transport.
func (p *Transport) SetLifetime(lifetime time.Duration) {
	if lifetime < 1 {
		return
	}
	p.maxLifetime = lifetime
}

// Send send a frame.
func (p *Transport) Send(frame frames.Frame, flush bool) (err error) {
	// defer func() {
	// 	// ensure frame done when send success.
	// 	if err == nil {
	// 		frame.Done()
	// 	}
	// }()
	if p == nil || p.conn == nil {
		err = errTransportClosed
		return
	}
	err = p.conn.Write(frame)
	if err != nil {
		return
	}
	if !flush {
		return
	}
	err = p.conn.Flush()
	return
}

// Flush flush all bytes in current connection.
func (p *Transport) Flush() (err error) {
	if p == nil || p.conn == nil {
		err = errTransportClosed
		return
	}
	err = p.conn.Flush()
	return
}

// Close close current transport.
func (p *Transport) Close() (err error) {
	p.once.Do(func() {
		err = p.conn.Close()
	})
	return
}

// ReadFirst reads first frame.
func (p *Transport) ReadFirst(ctx context.Context) (frame frames.Frame, err error) {
	select {
	case <-ctx.Done():
		err = ctx.Err()
	default:
		frame, err = p.conn.Read()
		if err != nil {
			err = fmt.Errorf("read first frame failed: %w", err)
		}
	}
	if err != nil {
		_ = p.Close()
	}
	return
}

func (p *Transport) loopReadBuffer(ctx context.Context, bf *buffer.Unbounded[frames.Frame], errCh chan<- error, done chan<- struct{}) {
	defer func() {
		done <- struct{}{}
	}()
	c := bf.Get()
	for next := range c {
		err := p.handler.HandleFrame(next)
		// err := p.DispatchFrame(ctx, next.(core.BufferedFrame))
		if err != nil {
			select {
			case errCh <- err:
			default:
			}
		}
		if bf.Load() == 0 {
			runtime.Gosched()
		}
	}
}

// Start start transport.
func (p *Transport) Start(ctx context.Context) error {
	defer func() {
		_ = p.Close()
	}()

	done := make(chan struct{}, 1)
	errChan := make(chan error, 1)

	var (
		framesBuffer = buffer.NewUnbounded[frames.Frame]()
	)

	defer func() {
		framesBuffer.Dispose()

		<-done
	}()

	go p.loopReadBuffer(ctx, framesBuffer, errChan, done)

	if p.isServer {
		first, err := p.ReadFirst(ctx)
		if err != nil {
			return fmt.Errorf("read first failed: %w", err)
		}
		if setup, ok := first.(*frames.Setup); ok {
			if err := p.handler.HandleFrame(setup); err != nil {
				return err
			}
		} else {
			return errors.New("expected first frame to be a setup frame")
		}
	}

	close(p.ready)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errChan:
			return fmt.Errorf("dispatch incoming frame failed: %w", err)
		default:
			f, err := p.conn.Read()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}

			framesBuffer.Put(f)
		}
	}
}

func (p *Transport) WaitUntilReady() {
	<-p.ready
}
