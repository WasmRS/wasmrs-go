package host

import (
	"context"

	"github.com/nanobus/iota/go/internal/frames"
	"github.com/nanobus/iota/go/invoke"
	"github.com/nanobus/iota/go/payload"
	"github.com/nanobus/iota/go/proxy"
	"github.com/nanobus/iota/go/rx"
	"github.com/nanobus/iota/go/rx/flux"
	"github.com/nanobus/iota/go/rx/mono"
)

func (i *Instance) ImportRequestResponse(namespace, operation string) uint32 {
	return invoke.ImportRequestResponse(namespace, operation)
}

func (i *Instance) ImportFireAndForget(namespace, operation string) uint32 {
	return invoke.ImportFireAndForget(namespace, operation)
}

func (i *Instance) ImportRequestStream(namespace, operation string) uint32 {
	return invoke.ImportRequestStream(namespace, operation)
}

func (i *Instance) ImportRequestChannel(namespace, operation string) uint32 {
	return invoke.ImportRequestChannel(namespace, operation)
}

func (i *Instance) RequestResponse(ctx context.Context, p payload.Payload) mono.Mono[payload.Payload] {
	return proxy.Mono(ctx, frames.RequestPayload{
		FrameType: frames.FrameTypeRequestResponse,
		StreamID:  i.getNextStreamID(),
		Metadata:  p.Metadata(),
		Data:      p.Data(),
		Complete:  true,
		InitialN:  1,
	}, i.SendFrame, i.registerStream)
}

func (i *Instance) FireAndForget(ctx context.Context, p payload.Payload) {
	i.SendFrame(&frames.RequestPayload{
		FrameType: frames.FrameTypeRequestFNF,
		StreamID:  i.getNextStreamID(),
		Metadata:  p.Metadata(),
		Data:      p.Data(),
		Complete:  true,
		InitialN:  1,
	})
}

func (i *Instance) RequestStream(ctx context.Context, p payload.Payload) flux.Flux[payload.Payload] {
	return proxy.Flux(ctx, frames.RequestPayload{
		FrameType: frames.FrameTypeRequestStream,
		StreamID:  i.getNextStreamID(),
		Metadata:  p.Metadata(),
		Data:      p.Data(),
		InitialN:  rx.RequestMax,
	}, nil, i.SendFrame, i.registerStream)
}

func (i *Instance) RequestChannel(ctx context.Context, p payload.Payload, in flux.Flux[payload.Payload]) flux.Flux[payload.Payload] {
	return proxy.Flux(ctx, frames.RequestPayload{
		FrameType: frames.FrameTypeRequestChannel,
		StreamID:  i.getNextStreamID(),
		Metadata:  p.Metadata(),
		Data:      p.Data(),
		InitialN:  rx.RequestMax,
	}, in, i.SendFrame, i.registerStream)
}

func (i *Instance) getNextStreamID() uint32 {
	streamID, _ := i.streamIDs.Next()
	return streamID
}
