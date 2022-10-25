package concat

import (
	"context"
	"encoding/binary"

	"github.com/nanobus/iota/go/msgpack"
	"github.com/nanobus/iota/go/wasmrs/invoke"
	"github.com/nanobus/iota/go/wasmrs/payload"
	"github.com/nanobus/iota/go/wasmrs/rx/mono"
	"github.com/nanobus/iota/go/wasmrs/transform"
)

var gCaller invoke.Caller

var (
	_opConcat uint32
)

func Initialize(caller invoke.Caller) {
	gCaller = caller
	_opConcat = caller.ImportRequestResponse("concat.v1", "concat")
}

func Concat(ctx context.Context, left, right string) mono.Mono[string] {
	request := Strings{
		Left:  left,
		Right: right,
	}
	payloadData, err := msgpack.ToBytes(&request)
	if err != nil {
		return mono.Error[string](err)
	}
	var metadata [8]byte
	binary.BigEndian.PutUint32(metadata[:], _opConcat)
	p := payload.New(payloadData, metadata[:])
	m := gCaller.RequestResponse(ctx, p)
	return mono.Map(m, transform.String.Decode)
}
