package proxy

import (
	"context"

	"github.com/nanobus/iota/go/payload"
)

type Stream interface {
	Context() context.Context
	StreamID() uint32
	OnNext(p payload.Payload)
	OnComplete()
	OnError(err error)
}

type streamKey struct{}

func FromContext(ctx context.Context) (Stream, bool) {
	iface := ctx.Value(streamKey{})
	if iface == nil {
		return nil, false
	}
	s, ok := ctx.Value(streamKey{}).(Stream)
	return s, ok
}

func WithContext(ctx context.Context, stream Stream) context.Context {
	return context.WithValue(ctx, streamKey{}, stream)
}
