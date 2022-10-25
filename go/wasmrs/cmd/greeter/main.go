package main

import (
	"context"

	_ "github.com/nanobus/iota/go/wasmrs/guest"
	"github.com/nanobus/iota/go/wasmrs/rx/mono"

	"github.com/nanobus/iota/go/wasmrs/example"
)

func main() {
	example.RegisterSayHello(SayHello)
}

func SayHello(ctx context.Context, firstName, lastName string) mono.Mono[string] {
	// return mono.Just("Hello")
	return mono.Just("Hello, " + firstName + " " + lastName)
}
