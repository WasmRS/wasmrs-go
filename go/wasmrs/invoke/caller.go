package invoke

import (
	"context"

	"github.com/nanobus/iota/go/wasmrs/payload"
	"github.com/nanobus/iota/go/wasmrs/rx/flux"
	"github.com/nanobus/iota/go/wasmrs/rx/mono"
)

type Caller interface {
	ImportRequestResponse(namespace, operation string) uint32
	ImportFireAndForget(namespace, operation string) uint32
	ImportRequestStream(namespace, operation string) uint32
	ImportRequestChannel(namespace, operation string) uint32

	RequestResponse(ctx context.Context, p payload.Payload) mono.Mono[payload.Payload]
	FireAndForget(ctx context.Context, p payload.Payload)
	RequestStream(ctx context.Context, p payload.Payload) flux.Flux[payload.Payload]
	RequestChannel(ctx context.Context, p payload.Payload, in flux.Flux[payload.Payload]) flux.Flux[payload.Payload]
}
