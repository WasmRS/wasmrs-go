package main

import (
	_ "github.com/nanobus/iota/go/wasmrs/guest"

	"github.com/nanobus/iota/go/wasmrs/example/concat"
	"github.com/nanobus/iota/go/wasmrs/example/greeter"
	"github.com/nanobus/iota/go/wasmrs/example/wordfrequency"
)

func main() {
	concat.Register()
	greeter.Register()
	wordfrequency.Register()
}
