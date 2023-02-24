package handler

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/nanobus/iota/go/internal/frames"
	"github.com/nanobus/iota/go/internal/socket"
	"github.com/nanobus/iota/go/invoke"
	"github.com/nanobus/iota/go/operations"
	"github.com/nanobus/iota/go/payload"
	"github.com/nanobus/iota/go/proxy"
	"github.com/nanobus/iota/go/rx"
	"github.com/nanobus/iota/go/rx/flux"
	"github.com/nanobus/iota/go/rx/mono"
)

type FrameSender func(f frames.Frame) error

type Handler struct {
	ctx          context.Context
	sendFrame    FrameSender
	streamIDs    uint64
	nextStreamID func() (id uint32, firstLoop bool)

	// Stream IDs on the guest MUST start at 1 and
	// increment by 2 sequentially, such as 1, 3, 5, 7, etc.
	guestStreams proxy.Lookup

	// Stream IDs on the host MUST start at 2 and
	// increment by 2 sequentially, such as 2, 4, 6, 8, etc.
	hostStreams proxy.Lookup

	// maxFrameSize       uint32
	// fragmentedPayloads map[uint32]fragmentedPayload

	importedRR   []invoke.RequestResponseHandler
	importedRFNF []invoke.FireAndForgetHandler
	importedRS   []invoke.RequestStreamHandler
	importedRC   []invoke.RequestChannelHandler

	opTable operations.Table
}

// type fragmentedPayload struct {
// 	frameType frames.FrameType
// 	initialN  uint32
// 	metadata  []byte
// 	data      []byte
// }

type Mode int

const (
	ClientMode Mode = iota
	ServerMode
)

func New(ctx context.Context, mode Mode) *Handler {
	h := Handler{
		ctx: ctx,
	}
	switch mode {
	case ClientMode:
		h.nextStreamID = (*socket.ClientStreamIDs)(&h.streamIDs).Next
	default:
		h.nextStreamID = (*socket.ServerStreamIDs)(&h.streamIDs).Next
	}
	return &h
}

func (i *Handler) SetFrameSender(sendFrame func(f frames.Frame) error) {
	i.sendFrame = sendFrame
}

func (i *Handler) Close() error {
	return nil
}

func (i *Handler) registerStream(s proxy.Stream) {
	if s.StreamID()&1 == 1 {
		i.guestStreams.Add(s)
	} else {
		i.hostStreams.Add(s)
	}
}

func (i *Handler) removeStream(streamID uint32) {
	if streamID&1 == 1 {
		i.guestStreams.Remove(streamID)
	} else {
		i.hostStreams.Remove(streamID)
	}
}

func (i *Handler) getStream(streamID uint32) (proxy.Stream, bool) {
	if streamID&1 == 1 {
		return i.guestStreams.Get(streamID)
	} else {
		return i.hostStreams.Get(streamID)
	}
}

func (i *Handler) SendFrame(f frames.Frame) error {
	return i.sendFrame(f)
}

func (i *Handler) HandleFrame(f frames.Frame) (err error) {
	var str proxy.Stream
	frameType := f.Type()
	streamID := f.GetStreamID()
	if frameType >= frames.FrameTypeRequestN {
		var ok bool
		str, ok = i.getStream(streamID)
		if !ok {
			// Return error
			if frameType != frames.FrameTypeRequestN {
				return fmt.Errorf("stream ID %d not found", streamID)
			}
			return
		}
	}

	switch v := f.(type) {
	case *frames.Setup:
		if i.opTable == nil {
			opers, err := operations.FromBytes(v.Data)
			if err != nil {
				return errors.New("could not read operations list")
			}
			i.opTable = opers
		}

	case *frames.RequestPayload:
		switch frameType {
		case frames.FrameTypeRequestResponse:
			// if i.checkFollows(&rr) {
			// 	// Will be processed under frames.FrameTypePayload
			// 	return
			// }
			go i.handleRequestResponse(i.ctx, streamID, v.Data, v.Metadata)

		case frames.FrameTypeRequestFNF:
			// if i.checkFollows(&rr) {
			// 	// Will be processed under frames.FrameTypePayload
			// 	return
			// }

			go i.handleFireAndForget(i.ctx, streamID, v.Data, v.Metadata)

		case frames.FrameTypeRequestStream:
			// if i.checkFollows(&rs) {
			// 	// Will be processed under frames.FrameTypePayload
			// 	return
			// }

			go i.handleRequestStream(i.ctx, streamID, v.Data, v.Metadata, v.InitialN)

		case frames.FrameTypeRequestChannel:
			// if i.checkFollows(&rc) {
			// 	// Will be processed under frames.FrameTypePayload
			// 	return
			// }

			go i.handleRequestChannel(i.ctx, streamID, v.Data, v.Metadata, v.InitialN)
		}

	case *frames.RequestN:
		s := str.(DoRequest)
		go s.DoRequest(int(v.N))

	case *frames.Cancel:
		str.OnComplete()
		i.removeStream(streamID)

	case *frames.Payload:
		// pl, ok := i.fragmentedPayloads[p.StreamID]
		// if ok {
		// 	pl = fragmentedPayload{
		// 		frameType: pl.frameType,
		// 		initialN:  pl.initialN,
		// 		metadata:  append(pl.metadata, p.Metadata...),
		// 		data:      append(pl.data, p.Data...),
		// 	}
		// } else {
		// 	pl = fragmentedPayload{
		// 		frameType: frames.FrameTypePayload,
		// 		metadata:  p.Metadata,
		// 		data:      p.Data,
		// 	}
		// }

		// if p.Follows {
		// 	i.fragmentedPayloads[p.StreamID] = pl
		// 	return
		// } else if ok {
		// 	delete(i.fragmentedPayloads, p.StreamID)
		// }

		// // Per https://rsocket.io/about/protocol#fragmentation-and-reassembly
		// // A reassembled payload is in `pl`.
		// //
		// switch pl.frameType {
		// case frames.FrameTypePayload:
		// 	if p.Next {
		// 		str.OnNext(payload.New(pl.data, pl.metadata))
		// 	}

		// case frames.FrameTypeRequestResponse:
		// 	i.handleRequestResponse(ctx, p.StreamID, pl.data, pl.metadata)

		// case frames.FrameTypeRequestFNF:
		// 	// TODO

		// case frames.FrameTypeRequestStream:
		// 	i.handleRequestStream(ctx, p.StreamID, pl.data, pl.metadata, pl.initialN)

		// case frames.FrameTypeRequestChannel:
		// 	i.handleRequestChannel(ctx, p.StreamID, pl.data, pl.metadata, pl.initialN)
		// }

		if v.Next {
			str.OnNext(payload.New(v.Data, v.Metadata))
		}

		if v.Complete {
			str.OnComplete()
			i.removeStream(streamID)
		}

	case *frames.Error:
		str.OnError(errors.New(v.Data))
		str.OnComplete()
		i.removeStream(streamID)
	}

	return nil
}

func (i *Handler) handleRequestResponse(ctx context.Context, streamID uint32, data, metadata []byte) {
	if !i.checkMetadata(streamID, metadata) {
		return
	}

	operationID := binary.BigEndian.Uint32(metadata)
	handler := invoke.GetRequestResponseHandler(operationID)
	if handler == nil {
		i.sendFrame(&frames.Error{
			StreamID: streamID,
			Data:     "not_found",
		})
		return
	}

	p := payload.New(data, metadata[8:])
	s := requestStream{ctx: ctx, streamID: streamID}
	ctx = proxy.WithContext(ctx, &s)
	handler(ctx, p).Subscribe(mono.Subscribe[payload.Payload]{
		OnSuccess: func(p payload.Payload) {
			i.sendFrame(&frames.Payload{
				StreamID: streamID,
				Metadata: p.Metadata(),
				Data:     p.Data(),
				Next:     true,
				Complete: true,
			})
		},
		OnError: func(err error) {
			i.sendFrame(&frames.Error{
				StreamID: streamID,
				Data:     err.Error(),
			})
		},
	})
}

func (i *Handler) handleFireAndForget(ctx context.Context, streamID uint32, data, metadata []byte) {
	if !i.checkMetadata(streamID, metadata) {
		return
	}

	operationID := binary.BigEndian.Uint32(metadata)
	handler := invoke.GetFireAndForgetHandler(operationID)
	if handler == nil {
		i.sendFrame(&frames.Error{
			StreamID: streamID,
			Data:     "not_found",
		})
		return
	}

	p := payload.New(data, metadata[8:])
	s := requestStream{ctx: ctx, streamID: streamID}
	i.registerStream(&s)
	ctx = proxy.WithContext(ctx, &s)
	handler(ctx, p)
}

func (i *Handler) handleRequestStream(ctx context.Context, streamID uint32, data, metadata []byte, initialN uint32) {
	if !i.checkMetadata(streamID, metadata) {
		return
	}

	operationID := binary.BigEndian.Uint32(metadata)
	handler := invoke.GetRequestStreamHandler(operationID)
	if handler == nil {
		i.sendFrame(&frames.Error{
			StreamID: streamID,
			Data:     "not_found",
		})
		return
	}

	p := payload.New(data, metadata)
	s := requestStream{ctx: ctx, streamID: streamID}
	i.registerStream(&s) // Need to register for RequestN frames
	// ctx := stream.WithContext(context.Background(), s)
	f := handler(ctx, p)
	f.Subscribe(flux.Subscribe[payload.Payload]{
		OnNext: func(p payload.Payload) {
			i.sendFrame(&frames.Payload{
				StreamID: streamID,
				Metadata: p.Metadata(),
				Data:     p.Data(),
				Next:     true,
			})
		},
		OnComplete: func() {
			i.sendFrame(&frames.Payload{
				StreamID: streamID,
				Complete: true,
			})
		},
		OnError: func(err error) {
			i.sendFrame(&frames.Error{
				StreamID: streamID,
				Data:     err.Error(),
			})
		},
		NoRequest: true,
	})
	sub := f.Subscription()
	s.sub = sub
	s.Request(int(initialN))
}

func (i *Handler) handleRequestChannel(ctx context.Context, streamID uint32, data, metadata []byte, initialN uint32) {
	if !i.checkMetadata(streamID, metadata) {
		return
	}

	operationID := binary.BigEndian.Uint32(metadata)
	handler := invoke.GetRequestChannelHandler(operationID)
	if handler == nil {
		i.sendFrame(&frames.Error{
			StreamID: streamID,
			Data:     "not_found",
		})
		return
	}

	p := payload.New(data, metadata)
	s := requestStream{streamID: streamID}
	i.registerStream(&s) // Need to register for RequestN frames
	in := flux.Create(func(sink flux.Sink[payload.Payload]) {
		s.sink = sink
		sink.OnSubscribe(flux.OnSubscribe{
			Request: func(n int) {
				i.sendFrame(&frames.RequestN{
					StreamID: streamID,
					N:        uint32(n),
				})
			},
			Cancel: func() {
				i.sendFrame(&frames.Cancel{
					StreamID: streamID,
				})
			},
		})
	})
	f := handler(ctx, p, in)
	f.Subscribe(flux.Subscribe[payload.Payload]{
		OnNext: func(p payload.Payload) {
			i.sendFrame(&frames.Payload{
				StreamID: streamID,
				Metadata: p.Metadata(),
				Data:     p.Data(),
				Next:     true,
			})
		},
		OnComplete: func() {
			i.sendFrame(&frames.Payload{
				StreamID: streamID,
				Complete: true,
			})
		},
		OnError: func(err error) {
			i.sendFrame(&frames.Error{
				StreamID: streamID,
				Data:     err.Error(),
			})
		},
		NoRequest: true,
	})
	sub := f.Subscription()
	s.sub = sub
	s.Request(int(initialN))
}

// func (i *Handler) handleRequestResponse(ctx context.Context, streamID uint32, data, metadata []byte) {
// 	if !i.checkMetadata(streamID, metadata) {
// 		return
// 	}

// 	fmt.Println("handleRequestResponse 1")

// 	operationID := binary.BigEndian.Uint32(metadata)
// 	parentStreamID := binary.BigEndian.Uint32(metadata[4:])
// 	if parentStreamID != 0 {
// 		if stream, ok := i.getStream(parentStreamID); ok {
// 			ctx = stream.Context()
// 		}
// 	}
// 	fmt.Println("handleRequestResponse 2", operationID)
// 	handler := i.getRequestResponseHandler(operationID)
// 	if handler == nil {
// 		fmt.Println(":(")
// 		i.SendFrame(&frames.Error{
// 			StreamID: streamID,
// 			Data:     "not_found",
// 		})
// 		return
// 	}
// 	fmt.Println("handleRequestResponse 3")
// 	p := payload.New(data, metadata)
// 	handler(ctx, p).Subscribe(mono.Subscribe[payload.Payload]{
// 		OnSuccess: func(p payload.Payload) {
// 			i.SendFrame(&frames.Payload{
// 				StreamID: streamID,
// 				Metadata: p.Metadata(),
// 				Data:     p.Data(),
// 				Next:     true,
// 				Complete: true,
// 			})
// 		},
// 		OnError: func(err error) {
// 			i.SendFrame(&frames.Error{
// 				StreamID: streamID,
// 				Data:     err.Error(),
// 			})
// 		},
// 	})
// }

// func (i *Handler) handleFireAndForget(ctx context.Context, streamID uint32, data, metadata []byte) {
// 	if !i.checkMetadata(streamID, metadata) {
// 		return
// 	}

// 	operationID := binary.BigEndian.Uint32(metadata)
// 	parentStreamID := binary.BigEndian.Uint32(metadata[4:])
// 	if parentStreamID != 0 {
// 		if stream, ok := i.getStream(parentStreamID); ok {
// 			ctx = stream.Context()
// 		}
// 	}
// 	handler := i.getFireAndForgetHandler(operationID)
// 	if handler == nil {
// 		i.SendFrame(&frames.Error{
// 			StreamID: streamID,
// 			Data:     "not_found",
// 		})
// 		return
// 	}
// 	p := payload.New(data, metadata)
// 	handler(ctx, p)
// }

// func (i *Handler) handleRequestStream(ctx context.Context, streamID uint32, data, metadata []byte, initialN uint32) {
// 	if !i.checkMetadata(streamID, metadata) {
// 		return
// 	}

// 	operationID := binary.BigEndian.Uint32(metadata)
// 	parentStreamID := binary.BigEndian.Uint32(metadata[4:])
// 	if parentStreamID != 0 {
// 		if stream, ok := i.getStream(parentStreamID); ok {
// 			ctx = stream.Context()
// 		}
// 	}
// 	handlerRS := i.getRequestStreamHandler(operationID)
// 	if handlerRS == nil {
// 		i.SendFrame(&frames.Error{
// 			StreamID: streamID,
// 			Data:     "not_found",
// 		})
// 		return
// 	}

// 	p := payload.New(data, metadata)
// 	s := requestStream{ctx: ctx, streamID: streamID}
// 	i.registerStream(&s) // Need to register for RequestN frames
// 	f := handlerRS(ctx, p)
// 	f.Subscribe(flux.Subscribe[payload.Payload]{
// 		OnNext: func(p payload.Payload) {
// 			i.SendFrame(&frames.Payload{
// 				StreamID: streamID,
// 				Metadata: p.Metadata(),
// 				Data:     p.Data(),
// 				Next:     true,
// 			})
// 		},
// 		OnComplete: func() {
// 			i.SendFrame(&frames.Payload{
// 				StreamID: streamID,
// 				Complete: true,
// 			})
// 		},
// 		OnError: func(err error) {
// 			i.SendFrame(&frames.Error{
// 				StreamID: streamID,
// 				Data:     err.Error(),
// 			})
// 		},
// 		NoRequest: true,
// 	})
// 	sub := f.Subscription()
// 	s.sub = sub
// 	s.Request(int(initialN))
// }

// func (i *Handler) handleRequestChannel(ctx context.Context, streamID uint32, data, metadata []byte, initialN uint32) {
// 	if !i.checkMetadata(streamID, metadata) {
// 		return
// 	}

// 	operationID := binary.BigEndian.Uint32(metadata)
// 	parentStreamID := binary.BigEndian.Uint32(metadata[4:])
// 	if parentStreamID != 0 {
// 		if stream, ok := i.getStream(parentStreamID); ok {
// 			ctx = stream.Context()
// 		}
// 	}
// 	handlerRC := i.getRequestChannelHandler(operationID)
// 	if handlerRC == nil {
// 		i.sendFrame(&frames.Error{
// 			StreamID: streamID,
// 			Data:     "not_found",
// 		})
// 		return
// 	}

// 	p := payload.New(data, metadata)
// 	s := requestStream{streamID: streamID}
// 	i.registerStream(&s) // Need to register for RequestN frames
// 	in := flux.Create(func(sink flux.Sink[payload.Payload]) {
// 		s.sink = sink
// 		sink.OnSubscribe(flux.OnSubscribe{
// 			Request: func(n int) {
// 				i.sendFrame(&frames.RequestN{
// 					StreamID: streamID,
// 					N:        uint32(n),
// 				})
// 			},
// 			Cancel: func() {
// 				i.sendFrame(&frames.Cancel{
// 					StreamID: streamID,
// 				})
// 			},
// 		})
// 	})
// 	f := handlerRC(ctx, p, in)
// 	f.Subscribe(flux.Subscribe[payload.Payload]{
// 		OnNext: func(p payload.Payload) {
// 			i.sendFrame(&frames.Payload{
// 				StreamID: streamID,
// 				Metadata: p.Metadata(),
// 				Data:     p.Data(),
// 				Next:     true,
// 			})
// 		},
// 		OnComplete: func() {
// 			i.sendFrame(&frames.Payload{
// 				StreamID: streamID,
// 				Complete: true,
// 			})
// 		},
// 		OnError: func(err error) {
// 			i.sendFrame(&frames.Error{
// 				StreamID: streamID,
// 				Data:     err.Error(),
// 			})
// 		},
// 		NoRequest: true,
// 	})
// 	sub := f.Subscription()
// 	s.sub = sub
// 	s.Request(int(initialN))
// }

func (i *Handler) SetRequestResponseHandler(index uint32, handler invoke.RequestResponseHandler) {
	for uint32(len(i.importedRR)) < index+1 {
		i.importedRR = append(i.importedRR, nil)
	}
	i.importedRR[index] = handler
}

func (i *Handler) getRequestResponseHandler(index uint32) invoke.RequestResponseHandler {
	if uint32(len(i.importedRR)) <= index {
		return nil
	}
	return i.importedRR[index]
}

func (i *Handler) SetFireAndForgetHandler(index uint32, handler invoke.FireAndForgetHandler) {
	for uint32(len(i.importedRFNF)) < index+1 {
		i.importedRFNF = append(i.importedRFNF, nil)
	}
	i.importedRFNF[index] = handler
}

func (i *Handler) getFireAndForgetHandler(index uint32) invoke.FireAndForgetHandler {
	if uint32(len(i.importedRFNF)) <= index {
		return nil
	}
	return i.importedRFNF[index]
}

func (i *Handler) SetRequestStreamHandler(index uint32, handler invoke.RequestStreamHandler) {
	for uint32(len(i.importedRS)) < index+1 {
		i.importedRS = append(i.importedRS, nil)
	}
	i.importedRS[index] = handler
}

func (i *Handler) getRequestStreamHandler(index uint32) invoke.RequestStreamHandler {
	if uint32(len(i.importedRS)) <= index {
		return nil
	}
	return i.importedRS[index]
}

func (i *Handler) SetRequestChannelHandler(index uint32, handler invoke.RequestChannelHandler) {
	for uint32(len(i.importedRC)) < index+1 {
		i.importedRC = append(i.importedRC, nil)
	}
	i.importedRC[index] = handler
}

func (i *Handler) getRequestChannelHandler(index uint32) invoke.RequestChannelHandler {
	if uint32(len(i.importedRC)) <= index {
		return nil
	}
	return i.importedRC[index]
}

func (i *Handler) checkMetadata(streamID uint32, metadata []byte) bool {
	if len(metadata) < 8 { // 48 before... but why?
		i.SendFrame(&frames.Error{
			StreamID: streamID,
			Data:     "Invalid metadata",
		})
		return false
	}

	return true
}

// func (i *Instance) checkFollows(requestFrame *frames.RequestPayload) bool {
// 	if requestFrame.Follows {
// 		i.fragmentedPayloads[requestFrame.StreamID] = fragmentedPayload{
// 			frameType: frames.FrameTypeRequestResponse,
// 			initialN:  requestFrame.InitialN,
// 			metadata:  requestFrame.Metadata,
// 			data:      requestFrame.Data,
// 		}
// 		return true
// 	}

// 	return false
// }

type DoRequest interface {
	DoRequest(n int)
}

type requestStream struct {
	ctx      context.Context
	streamID uint32
	flux.Sink[payload.Payload]
	sub  rx.Subscription
	sink flux.Sink[payload.Payload]
}

var _ = (proxy.Stream)((*requestStream)(nil))

func (r *requestStream) Context() context.Context {
	return r.ctx
}

func (r *requestStream) StreamID() uint32 {
	return r.streamID
}

func (r *requestStream) Request(n int) {
	r.sub.Request(n)
}

func (r *requestStream) DoRequest(n int) {
	r.sub.Request(n)
}

func (r *requestStream) OnNext(p payload.Payload) {
	r.sink.Next(p)
}

func (r *requestStream) OnComplete() {
	r.sink.Complete()
}

func (r *requestStream) OnError(err error) {
	r.sink.Error(err)
}
