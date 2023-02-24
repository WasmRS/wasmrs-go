package guest

import (
	"context"

	"github.com/nanobus/iota/go/internal/frames"
	"github.com/nanobus/iota/go/internal/socket"
	"github.com/nanobus/iota/go/invoke"
	"github.com/nanobus/iota/go/payload"
	"github.com/nanobus/iota/go/proxy"
	"github.com/nanobus/iota/go/rx"
	"github.com/nanobus/iota/go/rx/flux"
	"github.com/nanobus/iota/go/rx/mono"
)

var HostInvoker = &Caller{}

type Caller struct {
	streamIDs socket.ClientStreamIDs
}

func (i *Caller) ImportRequestResponse(namespace, operation string) uint32 {
	return invoke.ImportRequestResponse(namespace, operation)
}

func (i *Caller) ImportFireAndForget(namespace, operation string) uint32 {
	return invoke.ImportFireAndForget(namespace, operation)
}

func (i *Caller) ImportRequestStream(namespace, operation string) uint32 {
	return invoke.ImportRequestStream(namespace, operation)
}

func (i *Caller) ImportRequestChannel(namespace, operation string) uint32 {
	return invoke.ImportRequestChannel(namespace, operation)
}

func (i *Caller) RequestResponse(ctx context.Context, p payload.Payload) mono.Mono[payload.Payload] {
	return proxy.Mono(ctx, frames.RequestPayload{
		FrameType: frames.FrameTypeRequestResponse,
		StreamID:  i.getNextStreamID(),
		Metadata:  p.Metadata(),
		Data:      p.Data(),
		Complete:  true,
		InitialN:  1,
	}, sendFrame, registerStream)
}

func (i *Caller) FireAndForget(ctx context.Context, p payload.Payload) {
	sendFrame(&frames.RequestPayload{
		FrameType: frames.FrameTypeRequestFNF,
		StreamID:  i.getNextStreamID(),
		Metadata:  p.Metadata(),
		Data:      p.Data(),
		Complete:  true,
		InitialN:  1,
	})
}

func (i *Caller) RequestStream(ctx context.Context, p payload.Payload) flux.Flux[payload.Payload] {
	return proxy.Flux(ctx, frames.RequestPayload{
		FrameType: frames.FrameTypeRequestStream,
		StreamID:  i.getNextStreamID(),
		Metadata:  p.Metadata(),
		Data:      p.Data(),
		InitialN:  rx.RequestMax,
	}, nil, sendFrame, registerStream)
}

func (i *Caller) RequestChannel(ctx context.Context, p payload.Payload, in flux.Flux[payload.Payload]) flux.Flux[payload.Payload] {
	streamID := i.getNextStreamID()
	return proxy.Flux(ctx, frames.RequestPayload{
		FrameType: frames.FrameTypeRequestChannel,
		StreamID:  streamID,
		Metadata:  p.Metadata(),
		Data:      p.Data(),
		InitialN:  rx.RequestMax,
	}, in, sendFrame, registerStream)
}

func (i *Caller) getNextStreamID() uint32 {
	nextID, _ := i.streamIDs.Next()
	return nextID
}
