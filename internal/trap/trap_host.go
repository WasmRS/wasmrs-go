//go:build !purego && !appengine && !wasm && !tinygo.wasm && !wasi
// +build !purego,!appengine,!wasm,!tinygo.wasm,!wasi

package trap

func Trap() {
	panic("block")
}
