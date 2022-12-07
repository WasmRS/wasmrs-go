package mono

import (
	"github.com/nanobus/iota/go/rx"
)

type Void = Mono[struct{}]

var VoidVal Void = Just(struct{}{})

type Mono[T any] interface {
	Blockable[T]
	Subscribe(Subscribe[T]) Mono[T]
	Async()
	Notify(rx.FnFinally)
	Get() (T, error)
}

type Source[T any] func(Sink[T])

type Subscribe[T any] struct {
	OnSuccess rx.FnOnNext[T]
	OnError   rx.FnOnError
	OnRequest rx.FnOnCancel
	Finally   rx.FnFinally
}

type Sink[T any] interface {
	Success(T)
	Error(error)
}

func Error[T any](err error) Mono[T] {
	return errorSingle[T]{err}
}

type errorSingle[T any] struct {
	err error
}

func (e errorSingle[T]) Subscribe(sub Subscribe[T]) Mono[T] {
	if sub.OnError != nil {
		sub.OnError(e.err)
	}
	return e
}

func (e errorSingle[T]) Block() (T, error) {
	var dummy T
	return dummy, e.err
}

func (e errorSingle[T]) Get() (T, error) {
	var dummy T
	return dummy, e.err
}

func (e errorSingle[T]) Async() {
}

func (e errorSingle[T]) Notify(fn rx.FnFinally) {
	fn(rx.SignalError)
}

func Just[T any](value T) Mono[T] {
	return justSingle[T]{value}
}

type justSingle[T any] struct {
	value T
}

func (j justSingle[T]) Subscribe(sub Subscribe[T]) Mono[T] {
	if sub.OnSuccess != nil {
		sub.OnSuccess(j.value)
	}
	return j
}

func (j justSingle[T]) Block() (T, error) {
	return j.value, nil
}

func (j justSingle[T]) Get() (T, error) {
	return j.value, nil
}

func (j justSingle[T]) Async() {
}

func (j justSingle[T]) Notify(fn rx.FnFinally) {
	fn(rx.SignalComplete)
}

func Create[T any](source Source[T]) Mono[T] {
	return &mono[T]{
		source: source,
	}
}

type mono[T any] struct {
	source     Source[T]
	subscriber *subscriber[T]
	callbacks  []rx.FnFinally
}

func (m *mono[T]) Subscribe(sub Subscribe[T]) Mono[T] {
	wrapper := subscriber[T]{
		m:             m,
		doOnSuccess:   sub.OnSuccess,
		doOnError:     sub.OnError,
		doOnSubscribe: sub.OnRequest,
		doFinally:     sub.Finally,
	}
	m.subscriber = &wrapper
	m.source(&wrapper)
	return m
}

func (m *mono[T]) Async() {
}

func (m *mono[T]) Notify(fn rx.FnFinally) {
	// if m.subscriber == nil {
	// 	wrapper := subscriber[T]{}
	// 	m.subscriber = &wrapper
	// 	m.source(&wrapper)
	// }
	if m.subscriber != nil && m.subscriber.closed {
		fn(m.subscriber.signal)
		return
	}

	m.callbacks = append(m.callbacks, fn)
}

func (m *mono[T]) Get() (ret T, err error) {
	if !m.subscriber.closed {
		panic("mono: Get called before completion")
	}
	return m.subscriber.value, m.subscriber.err
}

type subscriber[T any] struct {
	closed bool
	value  T
	err    error
	signal rx.SignalType
	m      *mono[T]

	doOnSuccess   rx.FnOnNext[T]
	doOnError     rx.FnOnError
	doOnSubscribe rx.FnOnCancel
	doFinally     rx.FnFinally
}

func (s *subscriber[T]) Success(value T) {
	if s.closed {
		return
	}
	s.value = value
	s.closed = true
	if s.doOnSuccess != nil {
		s.doOnSuccess(value)
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
	for _, fn := range s.m.callbacks {
		fn(signal)
	}
}

func (s *subscriber[T]) Cancel() {
	if s.closed {
		return
	}
	s.closed = true
	s.finally(rx.SignalCancel)
}

func Map[S any | ~[]any, D any](mono Mono[S], tx rx.Transform[S, D]) Mono[D] {
	return &mapper[S, D]{
		mono: mono,
		tx:   tx,
	}
}

type mapper[S any, D any] struct {
	mono Mono[S]
	tx   rx.Transform[S, D]
}

func (b *mapper[S, D]) Subscribe(sub Subscribe[D]) Mono[D] {
	b.mono.Subscribe(Subscribe[S]{
		OnSuccess: func(value S) {
			converted, err := b.tx(value)
			if err != nil {
				if sub.OnError != nil {
					sub.OnError(err)
				}
			} else {
				if sub.OnSuccess != nil {
					sub.OnSuccess(converted)
				}
			}
		},
		OnError:   sub.OnError,
		OnRequest: sub.OnRequest,
		Finally:   sub.Finally,
	})
	return b
}

func (b *mapper[S, D]) Async() {
	b.mono.Async()
}

func (b *mapper[S, D]) Notify(fn rx.FnFinally) {
	b.mono.Notify(fn)
}

func (b *mapper[S, D]) Get() (D, error) {
	var d D
	v, err := b.mono.Get()
	if err != nil {
		return d, err
	}
	return b.tx(v)
}
