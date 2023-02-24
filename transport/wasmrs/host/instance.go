package host

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/tetratelabs/wazero/api"

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

type Instance struct {
	ctx       context.Context
	m         api.Module
	sendCh    chan frames.Frame
	recvCh    chan []byte
	sendFn    api.Function
	sendPtr   uint32
	sendSize  uint32
	recvPtr   uint32
	streamIDs socket.ServerStreamIDs

	// Stream IDs on the guest MUST start at 1 and
	// increment by 2 sequentially, such as 1, 3, 5, 7, etc.
	guestStreams proxy.Lookup

	// Stream IDs on the host MUST start at 2 and
	// increment by 2 sequentially, such as 2, 4, 6, 8, etc.
	hostStreams proxy.Lookup

	activeRequests atomic.Int64
	closing        atomic.Bool
	closed         chan struct{}
	once           sync.Once

	maxFrameSize       uint32
	fragmentedPayloads map[uint32]fragmentedPayload
	operations         operations.Table

	importedRR   []invoke.RequestResponseHandler
	importedRFNF []invoke.FireAndForgetHandler
	importedRS   []invoke.RequestStreamHandler
	importedRC   []invoke.RequestChannelHandler
}

type fragmentedPayload struct {
	frameType frames.FrameType
	initialN  uint32
	metadata  []byte
	data      []byte
}

type instanceKey struct{}

func NewInstance(ctx context.Context, m api.Module) (*Instance, error) {
	init := m.ExportedFunction("__wasmrs_init")
	if init == nil {
		return nil, errors.New("module does not export __wasmrs_init")
	}
	send := m.ExportedFunction("__wasmrs_send")
	if send == nil {
		return nil, errors.New("module does not export __wasmrs_send")
	}
	i := &Instance{
		ctx:                ctx,
		m:                  m,
		sendCh:             make(chan frames.Frame, 100),
		recvCh:             make(chan []byte, 100),
		sendFn:             send,
		maxFrameSize:       1024 * 1024,
		fragmentedPayloads: make(map[uint32]fragmentedPayload),
		sendSize:           16 * 1024,
		closed:             make(chan struct{}),
	}

	ctx = context.WithValue(ctx, instanceKey{}, i)
	_, err := init.Call(ctx, uint64(i.sendSize), uint64(16*1024), uint64(i.maxFrameSize))
	if err != nil {
		return nil, err
	}

	f := m.ExportedFunction("__wasmrs_op_list_request")
	_, err = f.Call(ctx)
	if err != nil {
		return nil, err
	}

	go i.recvLoop()
	go i.sendLoop()

	return i, nil
}

func (i *Instance) Close() error {
	i.once.Do(func() {
		i.closing.Store(true)

		if i.activeRequests.Load() > 0 {
			<-i.closed
		}

		close(i.sendCh)
		close(i.recvCh)
	})

	return nil
}

func (i *Instance) reduceActiveRequests() {
	if i.activeRequests.Add(-1) <= 0 && i.closing.Load() {
		close(i.closed)
	}
}

func (i *Instance) registerStream(s proxy.Stream) {
	if s.StreamID()&1 == 1 {
		i.guestStreams.Add(s)
	} else {
		i.hostStreams.Add(s)
	}
}

func (i *Instance) removeStream(streamID uint32) {
	if streamID&1 == 1 {
		i.guestStreams.Remove(streamID)
	} else {
		i.hostStreams.Remove(streamID)
	}
}

func (i *Instance) getStream(streamID uint32) (proxy.Stream, bool) {
	if streamID&1 == 1 {
		return i.guestStreams.Get(streamID)
	} else {
		return i.hostStreams.Get(streamID)
	}
}

func (i *Instance) SendFrame(f frames.Frame) error {
	// if frag, ok := f.(frames.Fragmentable); ok {
	// 	if frames := frag.Fragment(i.maxFrameSize); frames != nil {
	// 		for _, f := range frames {
	// 			i.sendCh <- f
	// 		}
	// 	} else {
	// 		i.sendCh <- f
	// 	}
	// } else {
	// 	i.sendCh <- f
	// }
	i.sendCh <- f
	return nil
}

func (i *Instance) setBuffers(sendPtr, recvPtr uint32) {
	i.sendPtr = sendPtr
	i.recvPtr = recvPtr
}

func (i *Instance) sendLoop() {
	var lengthBytes [4]byte
	ctx := context.WithValue(i.ctx, instanceKey{}, i)

	for f := range i.sendCh {
		mem := i.m.Memory()
		buf, _ := mem.Read(i.sendPtr, i.sendSize)
		byteCount := f.Size()

		// Encode length and frame data.
		binary.BigEndian.PutUint32(lengthBytes[:], byteCount)
		copy(buf[0:], lengthBytes[1:4])
		f.Encode(buf[3:])

		// Send frame data to guest.
		i.sendFn.Call(ctx, uint64(3+byteCount))
	}
}
func (i *Instance) Operations() operations.Table {
	return i.operations
}

func (i *Instance) opList(ctx context.Context, opPtr uint32, opSize uint32) {
	buf, _ := i.m.Memory().Read(opPtr, opSize)
	operations, _ := operations.FromBytes(buf)
	i.operations = operations
}

func (i *Instance) hostSend(ctx context.Context, recvPos uint32) {
	buffer, _ := i.m.Memory().Read(i.recvPtr, recvPos)

	for len(buffer) > 0 {
		var length [4]byte
		copy(length[1:4], buffer[0:3])
		buffer = buffer[3:]
		frameLength := binary.BigEndian.Uint32(length[0:4])
		buf := make([]byte, int(frameLength))
		copy(buf, buffer)
		i.recvCh <- buf
		buffer = buffer[frameLength:]
	}
}

func (i *Instance) recvLoop() {
	for buf := range i.recvCh {
		i.recvOne(buf)
	}
}

func (i *Instance) recvOne(data []byte) {
	ctx := context.Background()
	header := frames.ParseFrameHeader(data)
	data = data[frames.FrameHeaderLen:]

	var str proxy.Stream
	if header.Type() >= frames.FrameTypeRequestN {
		var ok bool
		str, ok = i.getStream(header.StreamID())
		if !ok {
			// Return error
			if header.Type() != frames.FrameTypeRequestN {
				fmt.Printf("Host: Stream %d not found\n", header.StreamID())
			}
			return
		}
	}

	switch header.Type() {
	case frames.FrameTypeSetup:

	case frames.FrameTypeRequestResponse:
		var rr frames.RequestPayload
		if err := rr.Decode(&header, data); err != nil {
			i.SendFrame(&frames.Error{
				StreamID: rr.StreamID,
				Data:     err.Error(),
			})
			return
		}
		if i.checkFollows(&rr) {
			// Will be processed under frames.FrameTypePayload
			return
		}

		i.activeRequests.Add(1)
		go i.handleRequestResponse(ctx, rr.StreamID, rr.Data, rr.Metadata)

	case frames.FrameTypeRequestFNF:
		var rr frames.RequestPayload
		if err := rr.Decode(&header, data); err != nil {
			i.SendFrame(&frames.Error{
				StreamID: rr.StreamID,
				Data:     err.Error(),
			})
			return
		}
		if i.checkFollows(&rr) {
			// Will be processed under frames.FrameTypePayload
			return
		}

		go i.handleFireAndForget(ctx, rr.StreamID, rr.Data, rr.Metadata)

	case frames.FrameTypeRequestStream:
		var rs frames.RequestPayload
		if err := rs.Decode(&header, data); err != nil {
			i.SendFrame(&frames.Error{
				StreamID: rs.StreamID,
				Data:     err.Error(),
			})
			return
		}
		if i.checkFollows(&rs) {
			// Will be processed under frames.FrameTypePayload
			return
		}

		i.activeRequests.Add(1)
		go i.handleRequestStream(ctx, rs.StreamID, rs.Data, rs.Metadata, rs.InitialN)

	case frames.FrameTypeRequestChannel:
		var rc frames.RequestPayload
		if err := rc.Decode(&header, data); err != nil {
			i.SendFrame(&frames.Error{
				StreamID: rc.StreamID,
				Data:     err.Error(),
			})
			return
		}
		if i.checkFollows(&rc) {
			// Will be processed under frames.FrameTypePayload
			return
		}

		i.activeRequests.Add(1)
		go i.handleRequestChannel(ctx, rc.StreamID, rc.Data, rc.Metadata, rc.InitialN)

	case frames.FrameTypeRequestN:
		var rn frames.RequestN
		if err := rn.Decode(&header, data); err != nil {
			i.SendFrame(&frames.Error{
				StreamID: rn.StreamID,
				Data:     err.Error(),
			})
			return
		}

		s := str.(DoRequest)
		go s.DoRequest(int(rn.N))

	case frames.FrameTypeCancel:
		str.OnComplete()
		i.removeStream(header.StreamID())

	case frames.FrameTypePayload:
		var p frames.Payload
		if err := p.Decode(&header, data); err != nil {
			return
		}

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

		if p.Next {
			str.OnNext(payload.New(p.Data, p.Metadata))
		}

		if p.Complete {
			str.OnComplete()
			i.removeStream(header.StreamID())
		}

	case frames.FrameTypeError:
		var p frames.Error
		if err := p.Decode(&header, data); err != nil {
			return
		}
		str.OnError(errors.New(p.Data))
		str.OnComplete()
		i.removeStream(header.StreamID())
	}
}

func (i *Instance) handleRequestResponse(ctx context.Context, streamID uint32, data, metadata []byte) {
	if !i.checkMetadata(streamID, metadata) {
		return
	}

	operationID := binary.BigEndian.Uint32(metadata)
	parentStreamID := binary.BigEndian.Uint32(metadata[4:])
	if parentStreamID != 0 {
		if stream, ok := i.getStream(parentStreamID); ok {
			ctx = stream.Context()
		}
	}
	handler := i.getRequestResponseHandler(operationID)
	if handler == nil {
		i.SendFrame(&frames.Error{
			StreamID: streamID,
			Data:     "not_found",
		})
		i.reduceActiveRequests()
		return
	}
	p := payload.New(data, metadata)
	handler(ctx, p).Subscribe(mono.Subscribe[payload.Payload]{
		OnSuccess: func(p payload.Payload) {
			i.SendFrame(&frames.Payload{
				StreamID: streamID,
				Metadata: p.Metadata(),
				Data:     p.Data(),
				Next:     true,
				Complete: true,
			})
			i.reduceActiveRequests()
		},
		OnError: func(err error) {
			i.SendFrame(&frames.Error{
				StreamID: streamID,
				Data:     err.Error(),
			})
			i.reduceActiveRequests()
		},
	})
}

func (i *Instance) handleFireAndForget(ctx context.Context, streamID uint32, data, metadata []byte) {
	if !i.checkMetadata(streamID, metadata) {
		return
	}

	operationID := binary.BigEndian.Uint32(metadata)
	parentStreamID := binary.BigEndian.Uint32(metadata[4:])
	if parentStreamID != 0 {
		if stream, ok := i.getStream(parentStreamID); ok {
			ctx = stream.Context()
		}
	}
	handler := i.getFireAndForgetHandler(operationID)
	if handler == nil {
		i.SendFrame(&frames.Error{
			StreamID: streamID,
			Data:     "not_found",
		})
		return
	}
	p := payload.New(data, metadata)
	handler(ctx, p)
}

func (i *Instance) handleRequestStream(ctx context.Context, streamID uint32, data, metadata []byte, initialN uint32) {
	if !i.checkMetadata(streamID, metadata) {
		return
	}

	operationID := binary.BigEndian.Uint32(metadata)
	parentStreamID := binary.BigEndian.Uint32(metadata[4:])
	if parentStreamID != 0 {
		if stream, ok := i.getStream(parentStreamID); ok {
			ctx = stream.Context()
		}
	}
	handlerRS := i.getRequestStreamHandler(operationID)
	if handlerRS == nil {
		i.SendFrame(&frames.Error{
			StreamID: streamID,
			Data:     "not_found",
		})
		return
	}

	p := payload.New(data, metadata)
	s := requestStream{ctx: ctx, streamID: streamID}
	f := handlerRS(ctx, p)
	f.Subscribe(flux.Subscribe[payload.Payload]{
		OnNext: func(p payload.Payload) {
			i.SendFrame(&frames.Payload{
				StreamID: streamID,
				Metadata: p.Metadata(),
				Data:     p.Data(),
				Next:     true,
			})
		},
		OnComplete: func() {
			i.SendFrame(&frames.Payload{
				StreamID: streamID,
				Complete: true,
			})
			i.reduceActiveRequests()
		},
		OnError: func(err error) {
			i.SendFrame(&frames.Error{
				StreamID: streamID,
				Data:     err.Error(),
			})
			i.reduceActiveRequests()
		},
		NoRequest: true,
	})
	sub := f.Subscription()
	s.sub = sub
	s.Request(int(initialN))
}

func (i *Instance) handleRequestChannel(ctx context.Context, streamID uint32, data, metadata []byte, initialN uint32) {
	if !i.checkMetadata(streamID, metadata) {
		return
	}

	operationID := binary.BigEndian.Uint32(metadata)
	parentStreamID := binary.BigEndian.Uint32(metadata[4:])
	if parentStreamID != 0 {
		if stream, ok := i.getStream(parentStreamID); ok {
			ctx = stream.Context()
		}
	}
	handlerRC := i.getRequestChannelHandler(operationID)
	if handlerRC == nil {
		i.SendFrame(&frames.Error{
			StreamID: streamID,
			Data:     "not_found",
		})
		i.reduceActiveRequests()
		return
	}

	p := payload.New(data, metadata)
	s := requestStream{streamID: streamID}
	i.registerStream(&s) // Need to register for RequestN frames
	in := flux.Create(func(sink flux.Sink[payload.Payload]) {
		s.sink = sink
		sink.OnSubscribe(flux.OnSubscribe{
			Request: func(n int) {
				i.SendFrame(&frames.RequestN{
					StreamID: streamID,
					N:        uint32(n),
				})
			},
			Cancel: func() {
				i.SendFrame(&frames.Cancel{
					StreamID: streamID,
				})
				i.reduceActiveRequests()
			},
		})
	})
	f := handlerRC(ctx, p, in)
	f.Subscribe(flux.Subscribe[payload.Payload]{
		OnNext: func(p payload.Payload) {
			i.SendFrame(&frames.Payload{
				StreamID: streamID,
				Metadata: p.Metadata(),
				Data:     p.Data(),
				Next:     true,
			})
		},
		OnComplete: func() {
			i.SendFrame(&frames.Payload{
				StreamID: streamID,
				Complete: true,
			})
			i.reduceActiveRequests()
		},
		OnError: func(err error) {
			i.SendFrame(&frames.Error{
				StreamID: streamID,
				Data:     err.Error(),
			})
			i.reduceActiveRequests()
		},
		NoRequest: true,
	})
	sub := f.Subscription()
	s.sub = sub
	s.Request(int(initialN))
}

func (i *Instance) SetRequestResponseHandler(index uint32, handler invoke.RequestResponseHandler) {
	for uint32(len(i.importedRR)) < index+1 {
		i.importedRR = append(i.importedRR, nil)
	}
	i.importedRR[index] = handler
}

func (i *Instance) getRequestResponseHandler(index uint32) invoke.RequestResponseHandler {
	if uint32(len(i.importedRR)) <= index {
		return nil
	}
	return i.importedRR[index]
}

func (i *Instance) SetFireAndForgetHandler(index uint32, handler invoke.FireAndForgetHandler) {
	for uint32(len(i.importedRFNF)) < index+1 {
		i.importedRFNF = append(i.importedRFNF, nil)
	}
	i.importedRFNF[index] = handler
}

func (i *Instance) getFireAndForgetHandler(index uint32) invoke.FireAndForgetHandler {
	if uint32(len(i.importedRFNF)) <= index {
		return nil
	}
	return i.importedRFNF[index]
}

func (i *Instance) SetRequestStreamHandler(index uint32, handler invoke.RequestStreamHandler) {
	for uint32(len(i.importedRS)) < index+1 {
		i.importedRS = append(i.importedRS, nil)
	}
	i.importedRS[index] = handler
}

func (i *Instance) getRequestStreamHandler(index uint32) invoke.RequestStreamHandler {
	if uint32(len(i.importedRS)) <= index {
		return nil
	}
	return i.importedRS[index]
}

func (i *Instance) SetRequestChannelHandler(index uint32, handler invoke.RequestChannelHandler) {
	for uint32(len(i.importedRC)) < index+1 {
		i.importedRC = append(i.importedRC, nil)
	}
	i.importedRC[index] = handler
}

func (i *Instance) getRequestChannelHandler(index uint32) invoke.RequestChannelHandler {
	if uint32(len(i.importedRC)) <= index {
		return nil
	}
	return i.importedRC[index]
}

func (i *Instance) checkMetadata(streamID uint32, metadata []byte) bool {
	if len(metadata) < 8 { // 48 before... but why?
		i.SendFrame(&frames.Error{
			StreamID: streamID,
			Data:     "Invalid metadata",
		})
		return false
	}

	return true
}

func (i *Instance) checkFollows(requestFrame *frames.RequestPayload) bool {
	if requestFrame.Follows {
		i.fragmentedPayloads[requestFrame.StreamID] = fragmentedPayload{
			frameType: frames.FrameTypeRequestResponse,
			initialN:  requestFrame.InitialN,
			metadata:  requestFrame.Metadata,
			data:      requestFrame.Data,
		}
		return true
	}

	return false
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
