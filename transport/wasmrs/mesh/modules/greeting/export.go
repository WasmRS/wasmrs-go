package greeting

import (
	"context"

	"github.com/nanobus/iota/go/invoke"
	"github.com/nanobus/iota/go/msgpack"
	"github.com/nanobus/iota/go/payload"
	"github.com/nanobus/iota/go/rx/mono"
	"github.com/nanobus/iota/go/transform"
)

type sayHelloHandler func(ctx context.Context, name string) mono.Mono[string]

func RegisterSayHello(fn sayHelloHandler) {
	invoke.ExportRequestResponse("greeting.v1", "sayHello", sayHelloWrapper(fn))
}

func sayHelloWrapper(handler sayHelloHandler) invoke.RequestResponseHandler {
	return func(ctx context.Context, p payload.Payload) mono.Mono[payload.Payload] {
		var request GreetingRequest
		decoder := msgpack.NewDecoder(p.Data())
		if err := request.Decode(&decoder); err != nil {
			return mono.Error[payload.Payload](err)
		}

		s := handler(ctx, request.Name)
		return mono.Map(s, transform.String.Encode)
	}
}
