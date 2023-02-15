package handler

import (
	"context"

	"github.com/nanobus/iota/go/internal/frames"
	"github.com/nanobus/iota/go/invoke"
	"github.com/nanobus/iota/go/operations"
	"github.com/nanobus/iota/go/payload"
	"github.com/nanobus/iota/go/proxy"
	"github.com/nanobus/iota/go/rx"
	"github.com/nanobus/iota/go/rx/flux"
	"github.com/nanobus/iota/go/rx/mono"
)

var _ = (invoke.Caller)((*Handler)(nil))

func (i *Handler) ImportRequestResponse(namespace, operation string) uint32 {
	for _, op := range i.opTable {
		if op.Direction == operations.Export &&
			op.Type == operations.RequestResponse &&
			op.Namespace == namespace &&
			op.Operation == operation {
			return op.Index
		}
	}
	return 0
}

func (i *Handler) ImportFireAndForget(namespace, operation string) uint32 {
	for _, op := range i.opTable {
		if op.Direction == operations.Export &&
			op.Type == operations.FireAndForget &&
			op.Namespace == namespace &&
			op.Operation == operation {
			return op.Index
		}
	}
	return 0
}

func (i *Handler) ImportRequestStream(namespace, operation string) uint32 {
	for _, op := range i.opTable {
		if op.Direction == operations.Export &&
			op.Type == operations.RequestStream &&
			op.Namespace == namespace &&
			op.Operation == operation {
			return op.Index
		}
	}
	return 0
}

func (i *Handler) ImportRequestChannel(namespace, operation string) uint32 {
	for _, op := range i.opTable {
		if op.Direction == operations.Export &&
			op.Type == operations.RequestChannel &&
			op.Namespace == namespace &&
			op.Operation == operation {
			return op.Index
		}
	}
	return 0
}

func (i *Handler) RequestResponse(ctx context.Context, p payload.Payload) mono.Mono[payload.Payload] {
	return proxy.Mono(ctx, frames.RequestPayload{
		FrameType: frames.FrameTypeRequestResponse,
		StreamID:  i.getNextStreamID(),
		Metadata:  p.Metadata(),
		Data:      p.Data(),
		Complete:  true,
		InitialN:  1,
	}, i.SendFrame, i.registerStream)
}

func (i *Handler) FireAndForget(ctx context.Context, p payload.Payload) {
	i.SendFrame(&frames.RequestPayload{
		FrameType: frames.FrameTypeRequestFNF,
		StreamID:  i.getNextStreamID(),
		Metadata:  p.Metadata(),
		Data:      p.Data(),
		Complete:  true,
		InitialN:  1,
	})
}

func (i *Handler) RequestStream(ctx context.Context, p payload.Payload) flux.Flux[payload.Payload] {
	return proxy.Flux(ctx, frames.RequestPayload{
		FrameType: frames.FrameTypeRequestStream,
		StreamID:  i.getNextStreamID(),
		Metadata:  p.Metadata(),
		Data:      p.Data(),
		InitialN:  rx.RequestMax,
	}, nil, i.SendFrame, i.registerStream)
}

func (i *Handler) RequestChannel(ctx context.Context, p payload.Payload, in flux.Flux[payload.Payload]) flux.Flux[payload.Payload] {
	return proxy.Flux(ctx, frames.RequestPayload{
		FrameType: frames.FrameTypeRequestChannel,
		StreamID:  i.getNextStreamID(),
		Metadata:  p.Metadata(),
		Data:      p.Data(),
		InitialN:  rx.RequestMax,
	}, in, i.SendFrame, i.registerStream)
}

func (i *Handler) getNextStreamID() uint32 {
	nextID, _ := i.nextStreamID()
	return nextID
}
