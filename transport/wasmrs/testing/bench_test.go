package testing

import (
	"context"
	"os"
	"testing"

	"github.com/nanobus/iota/go/transport/wasmrs/example"
	"github.com/nanobus/iota/go/transport/wasmrs/host"
)

// Prevent compiler optimization
var greetingMessage string

func BenchmarkInvoke(b *testing.B) {
	source, err := os.ReadFile("../cmd/greeter/main.wasm")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()

	h, err := host.New(ctx)
	if err != nil {
		panic(err)
	}
	module, err := h.Compile(ctx, source)
	if err != nil {
		panic(err)
	}
	inst, err := module.Instantiate(ctx)
	if err != nil {
		panic(err)
	}

	g := example.NewGreeting(inst)
	firstName := "John"
	lastName := "Doe"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		greetingMessage, err = g.SayHello(ctx, firstName, lastName).Block()
		if err != nil {
			b.Log(err)
			b.FailNow()
		}
	}
}

// Baseline: simple function call

func SayHello(ctx context.Context, firstName, lastName string) (string, error) {
	return "Hello, " + firstName + " " + lastName, nil
}

// Baseline
// BenchmarkInvoke-10    38391553     30.18 ns/op      16 B/op       1 allocs/op

// WasmRS
// 1000000000         0.001216 ns/op
// 1000000000         0.001464 ns/op
// 1000000000         0.001943 ns/op
