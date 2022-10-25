package concat

import (
	"context"

	"github.com/nanobus/iota/go/wasmrs/rx/mono"
)

func (c *Concat) Process(ctx context.Context) {
	c.Inputs.Strings.Subscribe(mono.Subscribe[Strings]{
		OnSuccess: func(value Strings) {
			c.Outputs.Value.Success(value.Left + " " + value.Right)
		},
		OnError: func(err error) {
			c.Outputs.Value.Error(err)
		},
	})
}
