package main

import (
	_ "github.com/nanobus/iota/go/transport/wasmrs/guest"

	"github.com/nanobus/iota/go/transport/wasmrs/example/concat"
	"github.com/nanobus/iota/go/transport/wasmrs/example/greeter"
	"github.com/nanobus/iota/go/transport/wasmrs/example/wordfrequency"
)

func main() {
	concat.Register()
	greeter.Register()
	wordfrequency.Register()
}
