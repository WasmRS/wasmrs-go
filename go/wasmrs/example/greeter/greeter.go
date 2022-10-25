package greeter

import (
	"context"

	"github.com/nanobus/iota/go/wasmrs/rx/mono"
)

func (g *Greeter) Process(ctx context.Context) {
	g.Inputs.Name.Subscribe(mono.Subscribe[string]{
		OnSuccess: func(value string) {
			g.Outputs.Message.Success("Hello, " + value)
		},
	})
}
