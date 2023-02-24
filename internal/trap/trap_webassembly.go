//go:build wasm || tinygo.wasm || wasi
// +build wasm tinygo.wasm wasi

package trap

//export llvm.trap
func Trap()
