//go:build !purego && !appengine && !wasm && !tinygo.wasm && !wasi
// +build !purego,!appengine,!wasm,!tinygo.wasm,!wasi

package proxy

import (
	"github.com/nanobus/iota/go/wasmrs/payload"
	"github.com/nanobus/iota/go/wasmrs/rx/mono"
)

func (s *streamMono) Block() (ret payload.Payload, err error) {
	// if s.subscriber != nil && s.subscriber.closed {
	// 	return s.subscriber.value, s.subscriber.err
	// }

	done := make(chan struct{})
	go func() {
		s.Subscribe(mono.Subscribe[payload.Payload]{
			OnSuccess: func(value payload.Payload) {
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
