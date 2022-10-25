package greeter

import (
	"context"

	"github.com/nanobus/iota/go/wasmrs/invoke"
	"github.com/nanobus/iota/go/wasmrs/payload"
	"github.com/nanobus/iota/go/wasmrs/rx"
	"github.com/nanobus/iota/go/wasmrs/rx/flux"
	"github.com/nanobus/iota/go/wasmrs/rx/mono"
	"github.com/nanobus/iota/go/wasmrs/transform"
)

const namespace = "greeter"

type (
	Greeter struct {
		Inputs  GreeterInputs
		Outputs GreeterOutputs
	}

	GreeterInputs struct {
		Name mono.Mono[string]
	}

	GreeterOutputs struct {
		Message mono.Sink[string]
	}
)

func Register() {
	invoke.ExportRequestChannel(namespace, "Greeter", greeterHandler)
}

func greeterHandler(ctx context.Context, pl payload.Payload, in flux.Flux[payload.Payload]) flux.Flux[payload.Payload] {
	return flux.Create(func(sink flux.Sink[payload.Payload]) {
		inName := mono.NewProcessor[string]()
		outMessage := mono.NewProcessor[string]()

		finalized := 0
		final := func(_ rx.SignalType) {
			finalized++
			if finalized == 1 {
				sink.Complete()
			}
		}

		outMessage.Subscribe(mono.Subscribe[string]{
			OnSuccess: func(next string) {
				p, err := transform.String.Encode(next)
				if err != nil {
					sink.Error(err)
					return
				}
				portIO := invoke.PortIO{
					Port:     "message",
					Next:     true,
					Complete: true,
				}
				p = payload.New(p.Data(), portIO.Encode())
				sink.Next(p)
			},
			Finally: final,
		})

		comp := Greeter{
			Inputs: GreeterInputs{
				Name: inName,
			},
			Outputs: GreeterOutputs{
				Message: outMessage,
			},
		}
		comp.Process(ctx)

		in.Subscribe(flux.Subscribe[payload.Payload]{
			OnNext: func(p payload.Payload) {
				var port invoke.PortIO
				if err := port.Decode(p.Metadata()); err != nil {
					sink.Error(err)
					return
				}

				switch port.Port {
				case "name":
					if port.Next {
						p := payload.New(p.Data())
						s, err := transform.String.Decode(p)
						if err != nil {
							sink.Error(err)
							return
						}
						inName.Success(s)
					}
				}
			},
		})
	})
}
