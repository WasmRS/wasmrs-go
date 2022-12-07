//go:build !purego && !appengine && !wasm && !tinygo.wasm && !wasi
// +build !purego,!appengine,!wasm,!tinygo.wasm,!wasi

package flux

import (
	"sync"

	"github.com/nanobus/iota/go/rx"
)

type Blockable[T any] interface {
	Block(Subscribe[T]) error
}

type mutex struct {
	sync.Mutex
}

func (s *subscriber[T]) doRequest(n int) {
	go func(n int) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.sub.Request(n)
	}(n)
}

func (f *flux[T]) Block(sub Subscribe[T]) (err error) {
	if f.subscriber.closed {
		return f.subscriber.err
	}

	return Block(Flux[T](f), sub)
}

func (b mapper[S, D]) Block(sub Subscribe[D]) (err error) {
	return Block(Flux[D](b), sub)
}

func Block[T any](f Flux[T], sub Subscribe[T]) (err error) {
	s := sub
	done := make(chan struct{}, 1)
	s.OnError = func(e error) {
		err = e
		if sub.OnError != nil {
			sub.OnError(e)
		}
	}
	s.Finally = func(signal rx.SignalType) {
		if sub.Finally != nil {
			sub.Finally(signal)
		}
		close(done)
	}
	go f.Subscribe(s)

	<-done
	return err
}
