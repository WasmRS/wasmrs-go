package main

import (
	"context"

	_ "github.com/nanobus/iota/go/wasmrs/guest"
	"github.com/nanobus/iota/go/wasmrs/rx/mono"

	"github.com/nanobus/iota/go/wasmrs/mesh/modules/concat"
)

func main() {
	concat.RegisterConcat(Concat)
}

func Concat(ctx context.Context, left, right string) mono.Mono[string] {
	return mono.Just(left + " " + right)
}
