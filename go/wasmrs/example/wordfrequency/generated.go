package wordfrequency

import (
	"context"

	"github.com/nanobus/iota/go/msgpack"
	"github.com/nanobus/iota/go/wasmrs/invoke"
	"github.com/nanobus/iota/go/wasmrs/payload"
	"github.com/nanobus/iota/go/wasmrs/rx"
	"github.com/nanobus/iota/go/wasmrs/rx/flux"
	"github.com/nanobus/iota/go/wasmrs/transform"
)

const namespace = "wordfrequency"

type (
	WordFrequency struct {
		Input  WordFrequencyInputs
		Output WordFrequencyOutputs
	}

	WordFrequencyInputs struct {
		Words flux.Flux[string]
	}

	WordFrequencyOutputs struct {
		Counts flux.Sink[WordCount]
	}
)

func Register() {
	invoke.ExportRequestChannel(namespace, "WordFrequency", wordFrequencyHandler)
}

func wordFrequencyHandler(ctx context.Context, pl payload.Payload, in flux.Flux[payload.Payload]) flux.Flux[payload.Payload] {
	return flux.Create(func(sink flux.Sink[payload.Payload]) {
		// Processors for each input and output.
		words := flux.NewProcessor[string]()
		counts := flux.NewProcessor[WordCount]()

		// Handle closing `sink` when all outputs are finalized.
		const numSinks = 1
		finalized := 0
		final := func(_ rx.SignalType) {
			finalized++
			if finalized == numSinks {
				sink.Complete()
			}
		}

		// Subscribe to counts and forward to `sink`
		// with PortIO metadata.
		counts.Subscribe(flux.Subscribe[WordCount]{
			OnNext: func(next WordCount) {
				p, err := transform.CodecEncode(&next)
				if err != nil {
					sink.Error(err)
					return
				}
				portIO := invoke.PortIO{
					Port: "counts",
					Next: true,
				}
				p = payload.New(p.Data(), portIO.Encode())
				sink.Next(p)
			},
			OnComplete: func() {
				pp := invoke.PortIO{
					Port:     "counts",
					Complete: true,
				}
				p := payload.New(nil, pp.Encode())
				sink.Next(p)
			},
			Finally: final,
		})

		// Create and invoke component.
		comp := WordFrequency{
			Input: WordFrequencyInputs{
				Words: words,
			},
			Output: WordFrequencyOutputs{
				Counts: counts,
			},
		}
		comp.Process(ctx)

		// Now that the component is ready,
		// Subscribe to `in`.
		in.Subscribe(flux.Subscribe[payload.Payload]{
			OnNext: func(p payload.Payload) {
				var port invoke.PortIO
				if err := port.Decode(p.Metadata()); err != nil {
					sink.Error(err)
					return
				}

				switch port.Port {
				case "words":
					if port.Next {
						next, err := transform.String.Decode(p)
						if err != nil {
							sink.Error(err)
							return
						}
						words.Next(next)
					}
					if port.Complete {
						words.Complete()
					}
				}
			},
			OnComplete: func() {
				words.Complete()
			},
		})
	})
}

type WordCount struct {
	Word  string `json:"word" msgpack:"word"`
	Count uint64 `json:"count" msgpack:"count"`
}

func (r *WordCount) Decode(decoder msgpack.Reader) error {
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
		case "word":
			r.Word, err = decoder.ReadString()
		case "count":
			r.Count, err = decoder.ReadUint64()
		default:
			err = decoder.Skip()
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *WordCount) Encode(encoder msgpack.Writer) error {
	if r == nil {
		encoder.WriteNil()
		return nil
	}
	encoder.WriteMapSize(2)
	encoder.WriteString("word")
	encoder.WriteString(r.Word)
	encoder.WriteString("count")
	encoder.WriteUint64(r.Count)
	return nil
}
