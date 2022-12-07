//go:build wasm || tinygo.wasm || wasi
// +build wasm tinygo.wasm wasi

package guest

//go:wasm-module wasmrs
//go:export __init_buffers
func initBuffers(guestBufferPtr uintptr, hostBufferPtr uintptr)

//go:wasm-module wasmrs
//go:export __op_list
func opList(opPtr uintptr, opSize uint32)

//go:wasm-module wasmrs
//go:export __send
func hostSend(nextSendPos uint32)
