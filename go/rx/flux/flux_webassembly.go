//go:build wasm || tinygo.wasm || wasi
// +build wasm tinygo.wasm wasi

package flux

type Blockable[T any] interface {
}

type mutex struct{}

func (s *subscriber[T]) doRequest(n int) {
	s.sub.Request(n)
}
