//go:build !purego && !appengine && !wasm && !tinygo.wasm && !wasi
// +build !purego,!appengine,!wasm,!tinygo.wasm,!wasi

package guest

// Stub host call for wasmrs:__init_buffers
func initBuffers(guestBufferPtr uintptr, hostBufferPtr uintptr) {}

// Stub host call for wasmrs:__op_list
func opList(opPtr uintptr, opSize uint32) {}

// Stub host call for wasmrs:__send
func hostSend(nextSendPos uint32) {}
