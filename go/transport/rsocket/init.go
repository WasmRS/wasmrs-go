package rsocket

import (
	"context"
	"os"

	"github.com/nanobus/iota/go/handler"
	"github.com/nanobus/iota/go/internal/frames"
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

func NewClient(ctx context.Context) (*Transport, error) {
	h := handler.New(ctx, handler.ClientMode)
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

func Connect() error {
	ctx := context.Background()
	t, err := NewClient(ctx)
	if err != nil {
		return err
	}

	return t.Start(ctx)
}

func getEnvOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	return value
}
