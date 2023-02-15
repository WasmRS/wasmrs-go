package rsocket

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/nanobus/iota/go/handler"
	"github.com/nanobus/iota/go/internal/frames"
	"github.com/nanobus/iota/go/invoke"
)

type tcpServerTransport struct {
	mu       sync.Mutex
	m        map[*Transport]struct{}
	lf       ListenerFactory
	hf       HandlerFactory
	l        net.Listener
	acceptor ServerTransportAcceptor
	done     chan struct{}
}

func (t *tcpServerTransport) Accept(acceptor ServerTransportAcceptor) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.acceptor = acceptor
}

func (t *tcpServerTransport) Close() (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	select {
	case <-t.done:
		// already closed
		break
	default:
		close(t.done)
		if t.l == nil {
			break
		}
		err = t.l.Close()
		for k := range t.m {
			_ = k.Close()
		}
		t.m = nil
	}
	return
}

func (t *tcpServerTransport) Listen(ctx context.Context, notifier chan<- bool) (err error) {
	t.l, err = t.lf(ctx)
	if err != nil {
		notifier <- false
		err = fmt.Errorf("listen tcp server failed: %w", err)
		return
	}

	defer func() {
		_ = t.Close()
	}()

	notifier <- true

	// daemon: close if ctx is done.
	go func() {
		select {
		case <-ctx.Done():
			_ = t.Close()
			break
		case <-t.done:
			break
		}
	}()

	// Start loop of accepting connections.
	for {
		var c net.Conn
		c, err = t.l.Accept()
		if err == io.EOF || isClosedErr(err) {
			err = nil
			break
		}
		if err != nil {
			err = fmt.Errorf("accept next conn failed: %w", err)
			continue
		}

		// Dispatch raw conn.
		h := handler.New(ctx, handler.ServerMode)
		tp := NewTransport(NewTCPConn(c), h, true)
		h.SetFrameSender(func(f frames.Frame) error {
			return tp.Send(f, true)
		})

		if t.putTransport(tp) {
			go tp.Start(ctx)
			go func() {
				tp.WaitUntilReady()
				t.acceptor(ctx, h, func(tp *Transport) {
					t.removeTransport(tp)
				})
			}()
		} else {
			_ = t.Close()
		}
	}
	return
}

func (t *tcpServerTransport) removeTransport(tp *Transport) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.m, tp)
}

func (t *tcpServerTransport) putTransport(tp *Transport) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	select {
	case <-t.done:
		// already closed
		return false
	default:
		if t.m == nil {
			return false
		}
		t.m[tp] = struct{}{}
		return true
	}
}

// NewTCPServerTransport creates a new server-side transport.
func NewTCPServerTransport(lf ListenerFactory, hf HandlerFactory) ServerTransport {
	return &tcpServerTransport{
		lf:       lf,
		hf:       hf,
		m:        make(map[*Transport]struct{}),
		done:     make(chan struct{}),
		acceptor: func(ctx context.Context, caller invoke.Caller, onClose func(*Transport)) {},
	}
}

// NewTCPListenerFactory creates a new server-side transport.
func NewTCPListenerFactory(network, addr string, tlsConfig *tls.Config) ListenerFactory {
	return func(ctx context.Context) (net.Listener, error) {
		var c net.ListenConfig
		l, err := c.Listen(ctx, network, addr)
		if err != nil {
			return nil, err
		}
		if tlsConfig == nil {
			return l, nil
		}
		return tls.NewListener(l, tlsConfig), nil
	}
}

// NewTCPClientTransport creates new transport.
func NewTCPClientTransport(c net.Conn, handler DuplexHandler) *Transport {
	return NewTransport(NewTCPConn(c), handler, false)
}

// NewConnWithAddr creates new connection.
func NewConnWithAddr(ctx context.Context, network, addr string, tlsConfig *tls.Config) (conn net.Conn, err error) {
	var dial net.Dialer
	conn, err = dial.DialContext(ctx, network, addr)
	if err != nil {
		return
	}
	if tlsConfig != nil {
		conn = tls.Client(conn, tlsConfig)
	}
	return
}
