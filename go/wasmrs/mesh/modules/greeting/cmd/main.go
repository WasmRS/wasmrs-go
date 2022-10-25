package main

import (
	"context"

	_ "github.com/nanobus/iota/go/wasmrs/guest"
	"github.com/nanobus/iota/go/wasmrs/rx/mono"

	"github.com/nanobus/iota/go/wasmrs/mesh/modules/greeting"
)

func main() {
	greeting.RegisterSayHello(SayHello)
}

func SayHello(ctx context.Context, name string) mono.Mono[string] {
	return mono.Just("Hello, " + name)
}
