package example

import (
	"context"
	"encoding/binary"

	"github.com/nanobus/iota/go/msgpack"
	"github.com/nanobus/iota/go/wasmrs/invoke"
	"github.com/nanobus/iota/go/wasmrs/payload"
	"github.com/nanobus/iota/go/wasmrs/rx/flux"
	"github.com/nanobus/iota/go/wasmrs/transform"
)

type Counter struct {
	caller     invoke.Caller
	opCountToN uint32
}

func NewCounter(caller invoke.Caller) *Counter {
	return &Counter{
		caller:     caller,
		opCountToN: caller.ImportRequestStream("host", "counter"),
	}
}

func (g *Counter) CountToN(ctx context.Context, to uint64) flux.Flux[string] {
	request := CounterRequest{
		To: to,
	}
	payloadData, err := msgpack.ToBytes(&request)
	if err != nil {
		return flux.Error[string](err)
	}
	var metadata [8]byte
	binary.BigEndian.PutUint32(metadata[:], g.opCountToN)
	p := payload.New(payloadData, metadata[:])

	f := g.caller.RequestStream(ctx, p)
	return flux.Map(f, transform.String.Decode)
}

type CounterRequest struct {
	To uint64 `json:"to" msgpack:"to"`
}

func (r *CounterRequest) Decode(decoder msgpack.Reader) error {
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
		case "to":
			r.To, err = decoder.ReadUint64()
		default:
			err = decoder.Skip()
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *CounterRequest) Encode(encoder msgpack.Writer) error {
	if r == nil {
		encoder.WriteNil()
		return nil
	}
	encoder.WriteMapSize(1)
	encoder.WriteString("to")
	encoder.WriteUint64(r.To)
	return nil
}

type counterHandler func(ctx context.Context, to uint64) flux.Flux[string]

func RegisterCounter(fn counterHandler) {
	invoke.ExportRequestStream("host", "counter", counterWrapper(fn))
}

func counterWrapper(handler counterHandler) invoke.RequestStreamHandler {
	return func(ctx context.Context, p payload.Payload) flux.Flux[payload.Payload] {
		var request CounterRequest
		decoder := msgpack.NewDecoder(p.Data())
		if err := request.Decode(&decoder); err != nil {
			return flux.Error[payload.Payload](err)
		}

		r := handler(ctx, request.To)
		return flux.Map(r, transform.String.Encode)
	}
}
