package main

import (
	"context"

	"github.com/nanobus/iota/go/rx"
	"github.com/nanobus/iota/go/rx/mono"

	"github.com/nanobus/iota/go/transport/wasmrs/example/concat"
	"github.com/nanobus/iota/go/transport/wasmrs/example/greeter"
)

func main() {
	ctx := context.Background()
	done := make(chan struct{}, 1)

	concatIn := mono.NewProcessor[concat.Strings]()
	concatOut := mono.NewProcessor[string]()

	greaterIn := mono.NewProcessor[string]()
	greeterOut := mono.NewProcessor[string]()

	comp1 := concat.Concat{
		Inputs: concat.ConcatInputs{
			Strings: concatIn,
		},
		Outputs: concat.ConcatOutputs{
			Value: concatOut,
		},
	}
	comp1.Process(ctx)

	comp2 := greeter.Greeter{
		Inputs: greeter.GreeterInputs{
			Name: greaterIn,
		},
		Outputs: greeter.GreeterOutputs{
			Message: greeterOut,
		},
	}
	comp2.Process(ctx)

	concatOut.Subscribe(mono.Subscribe[string]{
		OnSuccess: func(next string) {
			greaterIn.Success(next)
		},
	})

	greeterOut.Subscribe(mono.Subscribe[string]{
		OnSuccess: func(next string) {
			println(next)
		},
	})

	greeterOut.Notify(func(_ rx.SignalType) {
		close(done)
	})

	concatIn.Success(concat.Strings{
		Left:  "Foo",
		Right: "Bar",
	})

	<-done
}
