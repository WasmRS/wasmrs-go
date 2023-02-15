package rsocket

import (
	"context"
	"os"

	"github.com/nanobus/iota/go/handler"
	"github.com/nanobus/iota/go/internal/frames"
	"github.com/nanobus/iota/go/invoke"
)

func NewServer(acceptor ServerTransportAcceptor) ServerTransport {
	network := getEnvOrDefault("RSOCKET_NETWORK", "tcp")
	addr := getEnvOrDefault("RSOCKET_ADDRESS", ":7878")
	lf := NewTCPListenerFactory(network, addr, nil)
	server := NewTCPServerTransport(lf, func(ctx context.Context) (DuplexHandler, error) {
		return handler.New(ctx, handler.ServerMode), nil
	})
	server.Accept(acceptor)

	return server
}

func Handler(ctx context.Context) *handler.Handler {
	return handler.New(ctx, handler.ClientMode)
}

func NewClient(ctx context.Context, h *handler.Handler) (*Transport, error) {
	network := getEnvOrDefault("RSOCKET_NETWORK", "tcp")
	addr := getEnvOrDefault("RSOCKET_ADDRESS", "127.0.0.1:7878")
	conn, err := NewConnWithAddr(ctx, network, addr, nil)
	if err != nil {
		return nil, err
	}

	t := NewTCPClientTransport(conn, h)
	h.SetFrameSender(func(f frames.Frame) error {
		return t.Send(f, true)
	})

	return t, nil
}

func Connect(h *handler.Handler) error {
	ctx := context.Background()
	t, err := NewClient(ctx, h)
	if err != nil {
		return err
	}

	// Send operation list in setup frame.
	list := invoke.GetOperationsTable()
	payload := list.ToBytes()
	t.Send(&frames.Setup{
		MajorVersion: 0,
		MinorVersion: 2,
		Data:         payload,
	}, true)

	return t.Start(ctx)
}

func getEnvOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	return value
}
