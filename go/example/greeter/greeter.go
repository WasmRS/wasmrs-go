package greeter

import (
	"context"
	"encoding/binary"

	"github.com/nanobus/iota/go/invoke"
	"github.com/nanobus/iota/go/msgpack"
	"github.com/nanobus/iota/go/payload"
	"github.com/nanobus/iota/go/rx/mono"
	"github.com/nanobus/iota/go/transform"
)

type Greeter interface {
	SayHello(ctx context.Context, firstName, lastName string) mono.Mono[string]
}

type GreetingRequest struct {
	FirstName string `json:"firstName" msgpack:"firstName"`
	LastName  string `json:"lastName" msgpack:"lastName"`
}

func (r *GreetingRequest) Decode(decoder msgpack.Reader) error {
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
		case "firstName":
			r.FirstName, err = decoder.ReadString()
		case "lastName":
			r.LastName, err = decoder.ReadString()
		default:
			err = decoder.Skip()
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *GreetingRequest) Encode(encoder msgpack.Writer) error {
	if r == nil {
		encoder.WriteNil()
		return nil
	}
	encoder.WriteMapSize(2)
	encoder.WriteString("firstName")
	encoder.WriteString(r.FirstName)
	encoder.WriteString("lastName")
	encoder.WriteString(r.LastName)
	return nil
}

func RegisterGreeter(svc Greeter) {
	invoke.ExportRequestResponse("greeting", "sayHello", sayHelloWrapper(svc))
}

func sayHelloWrapper(svc Greeter) invoke.RequestResponseHandler {
	return func(ctx context.Context, p payload.Payload) mono.Mono[payload.Payload] {
		var request GreetingRequest
		decoder := msgpack.NewDecoder(p.Data())
		if err := request.Decode(&decoder); err != nil {
			return mono.Error[payload.Payload](err)
		}

		s := svc.SayHello(ctx, request.FirstName, request.LastName)
		return mono.Map(s, transform.String.Encode)
	}
}

type GreeterClient struct {
	caller     invoke.Caller
	opSayHello uint32
}

func NewGreeting(caller invoke.Caller) *GreeterClient {
	return &GreeterClient{
		caller:     caller,
		opSayHello: caller.ImportRequestResponse("greeting", "sayHello"),
	}
}

func (g *GreeterClient) SayHello(ctx context.Context, firstName, lastName string) mono.Mono[string] {
	request := GreetingRequest{
		FirstName: firstName,
		LastName:  lastName,
	}
	payloadData, err := msgpack.ToBytes(&request)
	if err != nil {
		return mono.Error[string](err)
	}
	var metadata [8]byte
	binary.BigEndian.PutUint32(metadata[:], g.opSayHello)
	p := payload.New(payloadData, metadata[:])
	m := g.caller.RequestResponse(ctx, p)
	return mono.Map(m, transform.String.Decode)
}
