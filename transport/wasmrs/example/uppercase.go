package example

import (
	"context"
	"encoding/binary"

	"github.com/nanobus/iota/go/invoke"
	"github.com/nanobus/iota/go/payload"
	"github.com/nanobus/iota/go/rx/flux"
	"github.com/nanobus/iota/go/transform"
)

type Uppercase struct {
	caller      invoke.Caller
	opUppercase uint32
}

func NewUppercase(caller invoke.Caller) *Uppercase {
	return &Uppercase{
		caller:      caller,
		opUppercase: caller.ImportRequestChannel("host", "counter"),
	}
}

func (g *Uppercase) Uppercase(ctx context.Context, in flux.Flux[string]) flux.Flux[string] {
	var metadata [8]byte
	binary.BigEndian.PutUint32(metadata[:], g.opUppercase)
	p := payload.New([]byte{}, metadata[:])
	f := g.caller.RequestChannel(ctx, p, flux.Map(in, transform.String.Encode))
	return flux.Map(f, transform.String.Decode)
}

type uppercaseHandler func(ctx context.Context, in flux.Flux[string]) flux.Flux[string]

func RegisterUppercase(fn uppercaseHandler) {
	invoke.ExportRequestChannel(
		"host", "counter", uppercaseWrapper(fn))
}

func uppercaseWrapper(handler uppercaseHandler) invoke.RequestChannelHandler {
	return func(ctx context.Context, p payload.Payload, in flux.Flux[payload.Payload]) flux.Flux[payload.Payload] {
		r := handler(ctx, flux.Map(in, transform.String.Decode))
		return flux.Map(r, transform.String.Encode)
	}
}
