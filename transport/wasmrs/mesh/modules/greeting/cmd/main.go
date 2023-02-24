package main

import (
	"context"

	"github.com/nanobus/iota/go/rx/mono"
	_ "github.com/nanobus/iota/go/transport/wasmrs/guest"

	"github.com/nanobus/iota/go/transport/wasmrs/mesh/modules/greeting"
)

func main() {
	greeting.RegisterSayHello(SayHello)
}

func SayHello(ctx context.Context, name string) mono.Mono[string] {
	return mono.Just("Hello, " + name)
}
