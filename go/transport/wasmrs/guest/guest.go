package guest

import (
	"context"
	"encoding/binary"
	"errors"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/nanobus/iota/go/internal/frames"
	"github.com/nanobus/iota/go/invoke"
	"github.com/nanobus/iota/go/payload"
	"github.com/nanobus/iota/go/proxy"
	"github.com/nanobus/iota/go/rx"
	"github.com/nanobus/iota/go/rx/flux"
	"github.com/nanobus/iota/go/rx/mono"
)

var (
	// Stream IDs on the guest MUST start at 1 and
	// increment by 2 sequentially, such as 1, 3, 5, 7, etc.
	guestStreams proxy.Lookup

	// Stream IDs on the host MUST start at 2 and
	// increment by 2 sequentially, such as 2, 4, 6, 8, etc.
	hostStreams proxy.Lookup

	// Send buffers
	guestBuffer []byte
	hostBuffer  []byte

	maxFrameSize       uint32
	fragmentedPayloads map[uint32]fragmentedPayload
)

type fragmentedPayload struct {
	frameType frames.FrameType
	initialN  uint32
	metadata  []byte
	data      []byte
}

//go:export __wasmrs_init
func Init(guestBufferSize uint32, hostBufferSize uint32, hostMaxFrameSize uint32) bool {
	//streams = make(map[uint32]proxy.Stream)
	fragmentedPayloads = make(map[uint32]fragmentedPayload)
	guestBuffer = make([]byte, guestBufferSize)
	hostBuffer = make([]byte, hostBufferSize)
	maxFrameSize = hostMaxFrameSize
	initBuffers(bytesToPointer(guestBuffer), bytesToPointer(hostBuffer))
	return true
}

// var buffers = sync.Pool{
// 	New: func() interface{} {
// 		return new(bytes.Buffer)
// 	},
// }

//go:export __wasmrs_op_list_request
func OperationList() {
	list := invoke.GetOperationsTable()
	payload := list.ToBytes()

	opList(bytesToPointer(payload), uint32(len(payload)))
}

//go:export __wasmrs_send
func GuestSend(endPos uint32) {
	ctx := context.Background()
	var lengthBuf [4]byte
	curPos := uint32(0)
	buf := guestBuffer

	for curPos < endPos {
		copy(lengthBuf[1:], buf)
		frameLen := binary.BigEndian.Uint32(lengthBuf[:])
		frameBuf := make([]byte, frameLen)
		copy(frameBuf, buf[3:])
		curPos += 3 + frameLen
		buf = buf[3+frameLen:]

		// lengthPos = 1 // Reset for next frame
		header := frames.ParseFrameHeader(frameBuf)
		buffer := frameBuf[frames.FrameHeaderLen:]

		var str proxy.Stream
		if header.Type() >= frames.FrameTypeRequestN {
			var ok bool
			str, ok = getStream(header.StreamID())
			if !ok {
				if header.Type() != frames.FrameTypeRequestN {
					println("Guest: Stream " + strconv.FormatUint(uint64(header.StreamID()), 10) + " not found")
				}
				// Return error
				continue
			}
		}

		switch header.Type() {
		// case frames.FrameTypeSetup:
		// 	// Not implemented

		case frames.FrameTypeRequestResponse:
			var rr frames.RequestPayload
			if err := rr.Decode(&header, buffer); err != nil {
				sendFrame(&frames.Error{
					StreamID: rr.StreamID,
					Data:     err.Error(),
				})
				continue
			}
			if checkFollows(&rr) {
				// Will be processed under frames.FrameTypePayload
				continue
			}

			handleRequestResponse(ctx, rr.StreamID, rr.Data, rr.Metadata)

		case frames.FrameTypeRequestFNF:
			var rr frames.RequestPayload
			if err := rr.Decode(&header, buffer); err != nil {
				sendFrame(&frames.Error{
					StreamID: rr.StreamID,
					Data:     err.Error(),
				})
				continue
			}
			if checkFollows(&rr) {
				// Will be processed under frames.FrameTypePayload
				continue
			}

			handleFireAndForget(ctx, rr.StreamID, rr.Data, rr.Metadata)

		case frames.FrameTypeRequestStream:
			var rs frames.RequestPayload
			if err := rs.Decode(&header, buffer); err != nil {
				sendFrame(&frames.Error{
					StreamID: rs.StreamID,
					Data:     err.Error(),
				})
				continue
			}
			if checkFollows(&rs) {
				// Will be processed under frames.FrameTypePayload
				continue
			}

			handleRequestStream(ctx, rs.StreamID, rs.Data, rs.Metadata, rs.InitialN)

		case frames.FrameTypeRequestChannel:
			var rc frames.RequestPayload
			if err := rc.Decode(&header, buffer); err != nil {
				sendFrame(&frames.Error{
					StreamID: rc.StreamID,
					Data:     err.Error(),
				})
				return
			}
			if checkFollows(&rc) {
				// Will be processed under frames.FrameTypePayload
				continue
			}

			handleRequestChannel(ctx, rc.StreamID, rc.Data, rc.Metadata, rc.InitialN)

		case frames.FrameTypeRequestN:
			var rn frames.RequestN
			if err := rn.Decode(&header, buffer); err != nil {
				sendFrame(&frames.Error{
					StreamID: rn.StreamID,
					Data:     err.Error(),
				})
				return
			}

			s := str.(DoRequest)
			s.DoRequest(int(rn.N))

		case frames.FrameTypeCancel:
			str.OnComplete()
			removeStream(header.StreamID())

		case frames.FrameTypePayload:
			var p frames.Payload
			if err := p.Decode(&header, buffer); err != nil {
				continue
			}

			// pl, ok := fragmentedPayloads[p.StreamID]
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

			pl := fragmentedPayload{
				frameType: frames.FrameTypePayload,
				metadata:  p.Metadata,
				data:      p.Data,
			}

			// if p.Follows {
			// 	fragmentedPayloads[p.StreamID] = pl
			// 	continue
			// } else if ok {
			// 	delete(fragmentedPayloads, p.StreamID)
			// }

			// Per https://rsocket.io/about/protocol#fragmentation-and-reassembly
			// A reassembled payload is in `pl`.
			//

			// switch pl.frameType {
			// case frames.FrameTypePayload:
			if p.Next {
				str.OnNext(payload.New(pl.data, pl.metadata))
			}

			// case frames.FrameTypeRequestResponse:
			// 	handleRequestResponse(ctx, p.StreamID, pl.data, pl.metadata)

			// case frames.FrameTypeRequestFNF:
			// 	// TODO

			// case frames.FrameTypeRequestStream:
			// 	handleRequestStream(ctx, p.StreamID, pl.data, pl.metadata, pl.initialN)

			// case frames.FrameTypeRequestChannel:
			// 	handleRequestChannel(ctx, p.StreamID, pl.data, pl.metadata, pl.initialN)
			// }

			if p.Complete {
				str.OnComplete()
				removeStream(header.StreamID())
			}

		case frames.FrameTypeError:
			var p frames.Error
			if err := p.Decode(&header, buffer); err != nil {
				continue
			}
			str.OnError(errors.New(p.Data))
			str.OnComplete()
			removeStream(header.StreamID())
		}
	}
}

func handleRequestResponse(ctx context.Context, streamID uint32, data, metadata []byte) {
	if !checkMetadata(streamID, metadata) {
		return
	}

	operationID := binary.BigEndian.Uint32(metadata)
	handler := invoke.GetRequestResponseHandler(operationID)
	if handler == nil {
		sendFrame(&frames.Error{
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
			sendFrame(&frames.Payload{
				StreamID: streamID,
				Metadata: p.Metadata(),
				Data:     p.Data(),
				Next:     true,
				Complete: true,
			})
		},
		OnError: func(err error) {
			sendFrame(&frames.Error{
				StreamID: streamID,
				Data:     err.Error(),
			})
		},
	})
}

func handleFireAndForget(ctx context.Context, streamID uint32, data, metadata []byte) {
	if !checkMetadata(streamID, metadata) {
		return
	}

	operationID := binary.BigEndian.Uint32(metadata)
	handler := invoke.GetFireAndForgetHandler(operationID)
	if handler == nil {
		sendFrame(&frames.Error{
			StreamID: streamID,
			Data:     "not_found",
		})
		return
	}

	p := payload.New(data, metadata[8:])
	s := requestStream{ctx: ctx, streamID: streamID}
	registerStream(&s)
	ctx = proxy.WithContext(ctx, &s)
	handler(ctx, p)
}

func handleRequestStream(ctx context.Context, streamID uint32, data, metadata []byte, initialN uint32) {
	if !checkMetadata(streamID, metadata) {
		return
	}

	operationID := binary.BigEndian.Uint32(metadata)
	handler := invoke.GetRequestStreamHandler(operationID)
	if handler == nil {
		sendFrame(&frames.Error{
			StreamID: streamID,
			Data:     "not_found",
		})
		return
	}

	p := payload.New(data, metadata)
	s := requestStream{ctx: ctx, streamID: streamID}
	registerStream(&s) // Need to register for RequestN frames
	// ctx := stream.WithContext(context.Background(), s)
	f := handler(ctx, p)
	f.Subscribe(flux.Subscribe[payload.Payload]{
		OnNext: func(p payload.Payload) {
			sendFrame(&frames.Payload{
				StreamID: streamID,
				Metadata: p.Metadata(),
				Data:     p.Data(),
				Next:     true,
			})
		},
		OnComplete: func() {
			sendFrame(&frames.Payload{
				StreamID: streamID,
				Complete: true,
			})
		},
		OnError: func(err error) {
			sendFrame(&frames.Error{
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

func handleRequestChannel(ctx context.Context, streamID uint32, data, metadata []byte, initialN uint32) {
	if !checkMetadata(streamID, metadata) {
		return
	}

	operationID := binary.BigEndian.Uint32(metadata)
	handler := invoke.GetRequestChannelHandler(operationID)
	if handler == nil {
		sendFrame(&frames.Error{
			StreamID: streamID,
			Data:     "not_found",
		})
		return
	}

	p := payload.New(data, metadata)
	s := requestStream{streamID: streamID}
	registerStream(&s) // Need to register for RequestN frames
	in := flux.Create(func(sink flux.Sink[payload.Payload]) {
		s.sink = sink
		sink.OnSubscribe(flux.OnSubscribe{
			Request: func(n int) {
				sendFrame(&frames.RequestN{
					StreamID: streamID,
					N:        uint32(n),
				})
			},
			Cancel: func() {
				sendFrame(&frames.Cancel{
					StreamID: streamID,
				})
			},
		})
	})
	f := handler(ctx, p, in)
	f.Subscribe(flux.Subscribe[payload.Payload]{
		OnNext: func(p payload.Payload) {
			sendFrame(&frames.Payload{
				StreamID: streamID,
				Metadata: p.Metadata(),
				Data:     p.Data(),
				Next:     true,
			})
		},
		OnComplete: func() {
			sendFrame(&frames.Payload{
				StreamID: streamID,
				Complete: true,
			})
		},
		OnError: func(err error) {
			sendFrame(&frames.Error{
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

func checkMetadata(streamID uint32, metadata []byte) bool {
	if len(metadata) < 8 { // 48 before but why?
		sendFrame(&frames.Error{
			StreamID: streamID,
			Data:     "Invalid metadata",
		})
		return false
	}

	return true
}

func checkFollows(requestFrame *frames.RequestPayload) bool {
	if requestFrame.Follows {
		fragmentedPayloads[requestFrame.StreamID] = fragmentedPayload{
			frameType: frames.FrameTypeRequestResponse,
			initialN:  requestFrame.InitialN,
			metadata:  requestFrame.Metadata,
			data:      requestFrame.Data,
		}
		return true
	}

	return false
}

func registerStream(s proxy.Stream) {
	if s.StreamID()&1 == 1 {
		guestStreams.Add(s)
	} else {
		hostStreams.Add(s)
	}
}

func removeStream(streamID uint32) {
	if streamID&1 == 1 {
		guestStreams.Remove(streamID)
	} else {
		hostStreams.Remove(streamID)
	}
}

func getStream(streamID uint32) (proxy.Stream, bool) {
	if streamID&1 == 1 {
		return guestStreams.Get(streamID)
	} else {
		return hostStreams.Get(streamID)
	}
}

// //go:inline
// func getStream(streamID uint32) (proxy.Stream, bool) {
// 	stream, ok := streams[streamID]
// 	return stream, ok
// }

// //go:inline
// func registerStream(s proxy.Stream) {
// 	streams[s.StreamID()] = s
// }

// //go:inline
// func removeStream(streamID uint32) {
// 	delete(streams, streamID)
// }

//go:inline
func sendFrame(f frames.Frame) error {
	// if f.Type() == frames.FrameTypePayload {
	// 	if frag, ok := f.(frames.Fragmentable); ok {
	// 		if frames := frag.Fragment(maxFrameSize); frames != nil {
	// 			for _, f := range frames {
	// 				doSendFrame(f)
	// 			}
	// 		} else {
	// 			doSendFrame(f)
	// 		}
	// 		return
	// 	}
	// }
	// if frag, ok := f.(frames.Fragmentable); ok {
	// 	if frames := frag.Fragment(maxFrameSize); frames != nil {
	// 		for _, f := range frames {
	// 			doSendFrame(f)
	// 		}
	// 	} else {
	// 		doSendFrame(f)
	// 	}
	// } else {
	// 	doSendFrame(f)
	// }
	doSendFrame(f)
	return nil
}

func doSendFrame(f frames.Frame) {
	byteLength := f.Size()
	var length [4]byte
	binary.BigEndian.PutUint32(length[0:4], byteLength)
	copy(hostBuffer[0:3], length[1:4])
	f.Encode(hostBuffer[3:])
	hostSend(byteLength + 3)
}

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

//go:inline
func bytesToPointer(s []byte) uintptr {
	//nolint
	return (*(*reflect.SliceHeader)(unsafe.Pointer(&s))).Data
}

//go:inline
func stringToPointer(s string) uintptr {
	//nolint
	return (*(*reflect.StringHeader)(unsafe.Pointer(&s))).Data
}
