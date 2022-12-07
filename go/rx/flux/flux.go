package flux

import (
	"github.com/nanobus/iota/go/rx"
)

type Flux[T any] interface {
	Blockable[T]
	Subscribe(Subscribe[T]) Flux[T]
	Async()
	Notify(rx.FnFinally)
	Subscription() rx.Subscription
}

type Source[T any] func(Sink[T])

type Subscribe[T any] struct {
	OnNext     rx.FnOnNext[T]
	OnComplete rx.FnOnComplete
	OnError    rx.FnOnError
	OnRequest  rx.FnOnSubscribe
	Finally    rx.FnFinally
	NoRequest  bool
}

type OnSubscribe struct {
	Request func(n int)
	Cancel  func()
}

type Sink[T any] interface {
	Next(T)
	Complete()
	Error(error)
	OnSubscribe(OnSubscribe)
}

func Error[T any](err error) Flux[T] {
	return errorFlow[T]{err}
}

type errorFlow[T any] struct {
	err error
}

func (e errorFlow[T]) Subscribe(sub Subscribe[T]) Flux[T] {
	if sub.OnError != nil {
		sub.OnError(e.err)
	}
	return e
}

func (e errorFlow[T]) Async() {}

func (e errorFlow[T]) Notify(fn rx.FnFinally) {
	fn(rx.SignalError)
}

func (e errorFlow[T]) Block(_ Subscribe[T]) error {
	return e.err
}

func (e errorFlow[T]) Subscription() rx.Subscription {
	return nil
}

func Create[T any](source Source[T]) Flux[T] {
	return &flux[T]{
		source: source,
	}
}

type flux[T any] struct {
	source     Source[T]
	subscriber *subscriber[T]
	callbacks  []rx.FnFinally
}

func (f *flux[T]) Subscribe(sub Subscribe[T]) Flux[T] {
	s := subscriber[T]{
		doOnComplete: sub.OnComplete,
		doOnError:    sub.OnError,
		doOnNext:     sub.OnNext,
		doOnRequest:  sub.OnRequest,
		doFinally:    sub.Finally,
		f:            f,
		noRequest:    sub.NoRequest,
	}
	f.subscriber = &s
	f.source(&s)
	s.handleOnSubscribe()
	return f
}

func (f *flux[T]) Subscription() rx.Subscription {
	return f.subscriber
}

func (f *flux[T]) Async() {}

func (f *flux[T]) Notify(fn rx.FnFinally) {
	if f.subscriber != nil && f.subscriber.closed {
		fn(f.subscriber.signal)
		return
	}
	f.callbacks = append(f.callbacks, fn)
}

type subscriber[T any] struct {
	closed    bool
	n         int
	err       error
	sub       OnSubscribe
	signal    rx.SignalType
	f         *flux[T]
	noRequest bool

	doOnNext     rx.FnOnNext[T]
	doOnComplete rx.FnOnComplete
	doOnError    rx.FnOnError
	doOnRequest  rx.FnOnSubscribe
	doFinally    rx.FnFinally

	mu mutex
}

func (s *subscriber[T]) Next(value T) {
	if s.closed {
		return
	}
	s.n--
	if s.doOnNext != nil {
		s.doOnNext(value)
	}

	if s.n == 0 {
		s.handleOnSubscribe()
	}
}

func (s *subscriber[T]) Complete() {
	if s.closed {
		return
	}
	s.closed = true
	if s.doOnComplete != nil {
		s.doOnComplete()
	}
	s.finally(rx.SignalComplete)
}

func (s *subscriber[T]) Error(err error) {
	if s.closed {
		return
	}
	s.err = err
	s.closed = true
	if s.doOnError != nil {
		s.doOnError(err)
	}
	s.finally(rx.SignalError)
}

func (s *subscriber[T]) finally(signal rx.SignalType) {
	s.signal = signal
	if s.doFinally != nil {
		s.doFinally(signal)
	}
	for _, fn := range s.f.callbacks {
		fn(signal)
	}
}

func (s *subscriber[T]) OnSubscribe(sub OnSubscribe) {
	if s.closed {
		return
	}
	s.sub = sub
	if s.n > 0 {
		s.doRequest(s.n)
	}
}

func (s *subscriber[T]) handleOnSubscribe() {
	if s.noRequest {
		return
	}
	if s.doOnRequest != nil {
		s.doOnRequest(s)
	} else {
		s.Request(rx.RequestMax)
	}
}

func (s *subscriber[T]) Request(n int) {
	s.n += n
	if s.sub.Request != nil {
		s.doRequest(n)
	}
}

func (s *subscriber[T]) Cancel() {
	s.closed = true
	if s.sub.Cancel != nil {
		s.sub.Cancel()
	}
	if s.doFinally != nil {
		s.doFinally(rx.SignalCancel)
	}
}

func (s *subscriber[T]) Close() {
	s.closed = true
}

func Map[S, D any](f Flux[S], tx rx.Transform[S, D]) Flux[D] {
	return mapper[S, D]{
		f:  f,
		tx: tx,
	}
}

type mapper[S, D any] struct {
	f  Flux[S]
	tx rx.Transform[S, D]
}

type closable[T any] interface {
	Close()
}

func (m mapper[S, D]) Subscribe(sub Subscribe[D]) Flux[D] {
	m.f.Subscribe(Subscribe[S]{
		OnNext: func(value S) {
			if converted, err := m.tx(value); err != nil {
				if close, ok := m.f.(closable[S]); ok {
					close.Close()
				}
				if sub.OnError != nil {
					sub.OnError(err)
				}
				if sub.Finally != nil {
					sub.Finally(rx.SignalError)
				}
			} else if sub.OnNext != nil {
				sub.OnNext(converted)
			}
		},
		OnComplete: sub.OnComplete,
		OnError:    sub.OnError,
		OnRequest:  sub.OnRequest,
		Finally:    sub.Finally,
		NoRequest:  sub.NoRequest,
	})
	return m
}

func (m mapper[S, D]) Async() {
	m.f.Async()
}

func (m mapper[S, D]) Notify(fn rx.FnFinally) {
	m.f.Notify(fn)
}

func (m mapper[S, D]) Subscription() rx.Subscription {
	return m.f.Subscription()
}
