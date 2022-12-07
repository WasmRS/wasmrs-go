package main

import (
	"context"

	"github.com/nanobus/iota/go/rx/mono"
	_ "github.com/nanobus/iota/go/transport/wasmrs/guest"

	"github.com/nanobus/iota/go/transport/wasmrs/mesh/modules/concat"
)

func main() {
	concat.RegisterConcat(Concat)
}

func Concat(ctx context.Context, left, right string) mono.Mono[string] {
	return mono.Just(left + " " + right)
}
