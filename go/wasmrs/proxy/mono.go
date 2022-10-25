package proxy

import (
	"context"

	"github.com/nanobus/iota/go/wasmrs/internal/frames"
	"github.com/nanobus/iota/go/wasmrs/payload"
	"github.com/nanobus/iota/go/wasmrs/rx"
	"github.com/nanobus/iota/go/wasmrs/rx/mono"
)

func Mono(ctx context.Context, request frames.RequestPayload, sendFrame func(frames.Frame), register func(Stream)) mono.Mono[payload.Payload] {
	m := mono.NewProcessor[payload.Payload]()
	ss := streamMono{
		ctx:       ctx,
		Processor: m,
		request:   request,
		sendFrame: sendFrame,
		register:  register,
	}
	return &ss
}

type streamMono struct {
	ctx context.Context
	mono.Processor[payload.Payload]

	subscribed bool
	request    frames.RequestPayload
	sendFrame  func(frames.Frame)
	register   func(Stream)
}

func (s *streamMono) Context() context.Context {
	return s.ctx
}

func (s *streamMono) StreamID() uint32 {
	return s.request.StreamID
}

func (s *streamMono) OnNext(p payload.Payload) {
	s.Processor.Success(p)
}

func (s *streamMono) OnComplete() {
	// Not implemented
}

func (s *streamMono) OnError(err error) {
	s.Processor.Error(err)
}

func (s *streamMono) Async() {
	// If treated like an awaitable and not subscribed,
	// then subscribe without callbacks.
	if !s.subscribed {
		s.Subscribe(mono.Subscribe[payload.Payload]{})
	}
}

func (s *streamMono) Notify(fn rx.FnFinally) {
	s.Processor.Notify(fn)
}

func (s *streamMono) Subscribe(sub mono.Subscribe[payload.Payload]) mono.Mono[payload.Payload] {
	s.subscribed = true
	s.register(s)
	s.Processor.Subscribe(sub)
	s.sendFrame(&s.request)
	return s
}
