//go:build wasm || tinygo.wasm || wasi
// +build wasm tinygo.wasm wasi

package mono

type Blockable[T any] interface {
}
