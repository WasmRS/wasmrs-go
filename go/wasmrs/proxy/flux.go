package proxy

import (
	"context"

	"github.com/nanobus/iota/go/wasmrs/internal/frames"
	"github.com/nanobus/iota/go/wasmrs/payload"
	"github.com/nanobus/iota/go/wasmrs/rx"
	"github.com/nanobus/iota/go/wasmrs/rx/flux"
)

func Flux(ctx context.Context, request frames.RequestPayload, in flux.Flux[payload.Payload], sendFrame func(frames.Frame), register func(Stream)) flux.Flux[payload.Payload] {
	p := flux.NewProcessor[payload.Payload]()
	ss := streamFlux{
		ctx:       ctx,
		Processor: p,
		request:   request,
		in:        in,
		sendFrame: sendFrame,
		register:  register,
		first:     true,
	}
	p.OnSubscribe(flux.OnSubscribe{
		Request: ss.Request,
		Cancel:  ss.Cancel,
	})
	return &ss
}

type streamFlux struct {
	ctx context.Context
	flux.Processor[payload.Payload]

	subscribed bool
	request    frames.RequestPayload
	in         flux.Flux[payload.Payload]
	sendFrame  func(frames.Frame)
	register   func(Stream)
	first      bool
	complete   bool
}

func (s *streamFlux) Context() context.Context {
	return s.ctx
}

func (s *streamFlux) StreamID() uint32 {
	return s.request.StreamID
}

func (s *streamFlux) OnNext(p payload.Payload) {
	s.Processor.Next(p)
}

func (s *streamFlux) OnComplete() {
	s.Processor.Complete()
	s.complete = true
}

func (s *streamFlux) OnError(err error) {
	s.Processor.Error(err)
}

func (s *streamFlux) Async() {
	// If treated like an awaitable and not subscribed,
	// then subscribe without callbacks.
	if !s.subscribed {
		s.Subscribe(flux.Subscribe[payload.Payload]{})
	}
}

func (s *streamFlux) Notify(fn rx.FnFinally) {
	s.Processor.Notify(fn)
}

func (s *streamFlux) Request(n int) {
	if s.complete {
		return
	}
	if s.first {
		s.request.InitialN = uint32(n)
		s.first = false
		s.sendFrame(&s.request)

		if s.in != nil {
			s.in.Subscribe(flux.Subscribe[payload.Payload]{
				OnNext: func(p payload.Payload) {
					s.sendFrame(&frames.Payload{
						StreamID: s.request.StreamID,
						Metadata: p.Metadata(),
						Data:     p.Data(),
						Next:     true,
					})
				},
				OnComplete: func() {
					s.sendFrame(&frames.Payload{
						StreamID: s.request.StreamID,
						Complete: true,
					})
				},
				OnError: func(err error) {
					s.sendFrame(&frames.Error{
						StreamID: s.request.StreamID,
						Data:     err.Error(),
					})
				},
				NoRequest: true,
			})
		}
	} else {
		s.sendFrame(&frames.RequestN{
			StreamID: s.request.StreamID,
			N:        uint32(n),
		})
	}
}

func (s *streamFlux) Cancel() {
	s.sendFrame(&frames.Cancel{
		StreamID: s.request.StreamID,
	})
}

func (s *streamFlux) DoRequest(n int) {
	sub := s.in.Subscription()
	sub.Request(n)
}

func (s *streamFlux) Subscribe(sub flux.Subscribe[payload.Payload]) flux.Flux[payload.Payload] {
	s.subscribed = true
	s.register(s)
	s.Processor.Subscribe(sub)
	return s
}
