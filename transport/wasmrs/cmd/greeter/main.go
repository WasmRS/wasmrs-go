package main

import (
	"context"

	"github.com/nanobus/iota/go/rx/mono"
	_ "github.com/nanobus/iota/go/transport/wasmrs/guest"

	"github.com/nanobus/iota/go/transport/wasmrs/example"
)

func main() {
	example.RegisterSayHello(SayHello)
}

func SayHello(ctx context.Context, firstName, lastName string) mono.Mono[string] {
	// return mono.Just("Hello")
	return mono.Just("Hello, " + firstName + " " + lastName)
}
