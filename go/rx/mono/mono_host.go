//go:build !purego && !appengine && !wasm && !tinygo.wasm && !wasi
// +build !purego,!appengine,!wasm,!tinygo.wasm,!wasi

package mono

type Blockable[T any] interface {
	Block() (T, error)
}

func (s *mono[T]) Block() (ret T, err error) {
	if s.subscriber != nil && s.subscriber.closed {
		return s.subscriber.value, s.subscriber.err
	}

	done := make(chan struct{})
	go func() {
		s.Subscribe(Subscribe[T]{
			OnSuccess: func(value T) {
				ret = value
				close(done)
			},
			OnError: func(e error) {
				err = e
				close(done)
			},
		})
	}()

	<-done
	return ret, err
}

func (b *mapper[S, D]) Block() (ret D, err error) {
	done := make(chan struct{})
	go func() {
		b.Subscribe(Subscribe[D]{
			OnSuccess: func(value D) {
				ret = value
				close(done)
			},
			OnError: func(e error) {
				err = e
				close(done)
			},
		})
	}()

	<-done
	return ret, err
}
