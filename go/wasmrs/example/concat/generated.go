package concat

import (
	"context"

	"github.com/nanobus/iota/go/msgpack"
	"github.com/nanobus/iota/go/wasmrs/invoke"
	"github.com/nanobus/iota/go/wasmrs/payload"
	"github.com/nanobus/iota/go/wasmrs/rx"
	"github.com/nanobus/iota/go/wasmrs/rx/flux"
	"github.com/nanobus/iota/go/wasmrs/rx/mono"
	"github.com/nanobus/iota/go/wasmrs/transform"
)

const namespace = "concat"

type (
	Concat struct {
		Inputs  ConcatInputs
		Outputs ConcatOutputs
	}

	ConcatInputs struct {
		Strings mono.Mono[Strings]
	}

	ConcatOutputs struct {
		Value mono.Sink[string]
	}
)

func Register() {
	invoke.ExportRequestChannel(namespace, "Concat", concatHandler)
}

func concatHandler(ctx context.Context, pl payload.Payload, in flux.Flux[payload.Payload]) flux.Flux[payload.Payload] {
	return flux.Create(func(sink flux.Sink[payload.Payload]) {
		strings := mono.NewProcessor[Strings]()
		joined := mono.NewProcessor[string]()

		finalized := 0
		final := func(_ rx.SignalType) {
			finalized++
			if finalized == 1 {
				sink.Complete()
			}
		}

		joined.Subscribe(mono.Subscribe[string]{
			OnSuccess: func(next string) {
				p, err := transform.String.Encode(next)
				if err != nil {
					sink.Error(err)
					return
				}
				portIO := invoke.PortIO{
					Port:     "value",
					Next:     true,
					Complete: true,
				}
				p = payload.New(p.Data(), portIO.Encode())
				sink.Next(p)
			},
			Finally: final,
		})

		comp := Concat{
			Inputs: ConcatInputs{
				Strings: strings,
			},
			Outputs: ConcatOutputs{
				Value: joined,
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
				case "strings":
					if port.Next {
						p := payload.New(p.Data())
						var s Strings
						err := transform.CodecDecode(p, &s)
						if err != nil {
							sink.Error(err)
							return
						}
						strings.Success(s)
					}
				}
			},
		})
	})
}

type Strings struct {
	Left  string `json:"left" msgpack:"left"`
	Right string `json:"right" msgpack:"right"`
}

func (s *Strings) Decode(decoder msgpack.Reader) error {
	numFields, err := decoder.ReadMapSize()
	if err != nil {
		return err
	}

	for numFields > 0 {
		numFields--
		field, err := decoder.ReadString()
		if err != nil {
			return err
		}
		switch field {
		case "left":
			s.Left, err = decoder.ReadString()
		case "right":
			s.Right, err = decoder.ReadString()
		default:
			err = decoder.Skip()
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Strings) Encode(encoder msgpack.Writer) error {
	if s == nil {
		encoder.WriteNil()
		return nil
	}
	encoder.WriteMapSize(2)
	encoder.WriteString("left")
	encoder.WriteString(s.Left)
	encoder.WriteString("right")
	encoder.WriteString(s.Right)
	return nil
}
