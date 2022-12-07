package greeting

import (
	"context"
	"encoding/binary"

	"github.com/nanobus/iota/go/invoke"
	"github.com/nanobus/iota/go/msgpack"
	"github.com/nanobus/iota/go/payload"
	"github.com/nanobus/iota/go/rx/mono"
	"github.com/nanobus/iota/go/transform"
)

var gCaller invoke.Caller

var (
	_opSayHello uint32
)

func Initialize(caller invoke.Caller) {
	gCaller = caller
	_opSayHello = caller.ImportRequestResponse("greeting.v1", "sayHello")
}

func SayHello(ctx context.Context, name string) mono.Mono[string] {
	request := GreetingRequest{
		Name: name,
	}
	payloadData, err := msgpack.ToBytes(&request)
	if err != nil {
		return mono.Error[string](err)
	}
	var metadata [8]byte
	binary.BigEndian.PutUint32(metadata[:], _opSayHello)
	p := payload.New(payloadData, metadata[:])
	m := gCaller.RequestResponse(ctx, p)
	return mono.Map(m, transform.String.Decode)
}
