package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nanobus/iota/go/wasmrs/example/concat"
	"github.com/nanobus/iota/go/wasmrs/host"
	"github.com/nanobus/iota/go/wasmrs/invoke"
	"github.com/nanobus/iota/go/wasmrs/payload"
	"github.com/nanobus/iota/go/wasmrs/rx"
	"github.com/nanobus/iota/go/wasmrs/rx/flux"
	"github.com/nanobus/iota/go/wasmrs/transform"
)

func main() {
	ctx := context.Background()
	inst, err := loadInstance(ctx)
	if err != nil {
		panic(err)
	}
	done := make(chan struct{}, 1)

	strs := concat.Strings{
		Left:  "Foo",
		Right: "Bar",
	}
	outConcat := invokeConcat(strs, inst, ctx)
	inGreeting, outGreeting := invokeGreet(inst, ctx)

	outConcat.Subscribe(flux.Subscribe[payload.Payload]{
		OnNext: func(p payload.Payload) {
			var portIO invoke.PortIO
			if err := portIO.Decode(p.Metadata()); err != nil {
				return
			}

			switch portIO.Port {
			case "value":
				// Forward to next component.
				portIO.Port = "name"
				inGreeting.Next(payload.New(p.Data(), portIO.Encode()))
			}
		},
	})

	outGreeting.Subscribe(flux.Subscribe[payload.Payload]{
		OnNext: func(p payload.Payload) {
			var portIO invoke.PortIO
			if err := portIO.Decode(p.Metadata()); err != nil {
				return
			}

			switch portIO.Port {
			case "message":
				value, err := transform.String.Decode(payload.New(p.Data()))
				if err != nil {
					return
				}
				fmt.Println(value)
			}
		},
		Finally: func(_ rx.SignalType) {
			close(done)
		},
	})

	<-done
}

func invokeConcat(strs concat.Strings, inst *host.Instance, ctx context.Context) flux.Flux[payload.Payload] {
	encoded, err := transform.CodecEncode(&strs)
	if err != nil {
		return flux.Error[payload.Payload](err)
	}

	portIO := invoke.PortIO{
		Port:     "strings",
		Next:     true,
		Complete: true,
	}
	var metadata [8]byte
	in1 := flux.Create(func(sink flux.Sink[payload.Payload]) {
		sink.Next(payload.New(encoded.Data(), portIO.Encode()))
		sink.Complete()
	})

	return inst.RequestChannel(ctx, payload.New([]byte{}, metadata[:]), in1)
}

func invokeGreet(inst *host.Instance, ctx context.Context) (flux.Processor[payload.Payload], flux.Flux[payload.Payload]) {
	var metadata [8]byte
	in := flux.NewProcessor[payload.Payload]()
	out := inst.RequestChannel(ctx, payload.New([]byte{}, metadata[:]), in)

	return in, out
}

func loadInstance(ctx context.Context) (*host.Instance, error) {
	source, err := os.ReadFile("cmd/fbp/guest/main.wasm")
	if err != nil {
		return nil, err
	}

	h, err := host.New(ctx)
	if err != nil {
		return nil, err
	}
	module, err := h.Compile(ctx, source)
	if err != nil {
		return nil, err
	}
	inst, err := module.Instantiate(ctx)
	if err != nil {
		return nil, err
	}
	return inst, nil
}
