package customer

import (
	"context"

	"github.com/nanobus/iota/go/msgpack"
	"github.com/nanobus/iota/go/wasmrs/invoke"
	"github.com/nanobus/iota/go/wasmrs/payload"
	"github.com/nanobus/iota/go/wasmrs/rx/mono"
	"github.com/nanobus/iota/go/wasmrs/transform"
)

type newCustomerHandler func(ctx context.Context, customer *Customer) mono.Mono[string]

func RegisterNewCustomer(fn newCustomerHandler) {
	invoke.ExportRequestResponse("flow.v1", "flow", newCustomerWrapper(fn))
}

func newCustomerWrapper(handler newCustomerHandler) invoke.RequestResponseHandler {
	return func(ctx context.Context, p payload.Payload) mono.Mono[payload.Payload] {
		var request Customer
		decoder := msgpack.NewDecoder(p.Data())
		if err := request.Decode(&decoder); err != nil {
			return mono.Error[payload.Payload](err)
		}

		s := handler(ctx, &request)
		return mono.Map(s, transform.String.Encode)
	}
}
