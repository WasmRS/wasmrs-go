package concat

import (
	"context"

	"github.com/nanobus/iota/go/invoke"
	"github.com/nanobus/iota/go/msgpack"
	"github.com/nanobus/iota/go/payload"
	"github.com/nanobus/iota/go/rx/mono"
	"github.com/nanobus/iota/go/transform"
)

type concatHandler func(ctx context.Context, left, right string) mono.Mono[string]

func RegisterConcat(fn concatHandler) {
	invoke.ExportRequestResponse("concat.v1", "concat", sayHelloWrapper(fn))
}

func sayHelloWrapper(handler concatHandler) invoke.RequestResponseHandler {
	return func(ctx context.Context, p payload.Payload) mono.Mono[payload.Payload] {
		var request Strings
		decoder := msgpack.NewDecoder(p.Data())
		if err := request.Decode(&decoder); err != nil {
			return mono.Error[payload.Payload](err)
		}

		s := handler(ctx, request.Left, request.Right)
		return mono.Map(s, transform.String.Encode)
	}
}
