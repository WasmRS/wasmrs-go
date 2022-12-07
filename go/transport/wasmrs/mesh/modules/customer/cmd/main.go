package main

import (
	"context"

	"github.com/nanobus/iota/go/rx/mono"
	"github.com/nanobus/iota/go/transport/wasmrs/await"
	"github.com/nanobus/iota/go/transport/wasmrs/flow"
	"github.com/nanobus/iota/go/transport/wasmrs/guest"

	"github.com/nanobus/iota/go/transport/wasmrs/mesh/modules/concat"
	"github.com/nanobus/iota/go/transport/wasmrs/mesh/modules/customer"
	"github.com/nanobus/iota/go/transport/wasmrs/mesh/modules/greeting"
)

func main() {
	concat.Initialize(guest.HostInvoker)
	greeting.Initialize(guest.HostInvoker)
	customer.RegisterNewCustomer(NewCustomer)
}

func NewCustomer(ctx context.Context, customer *customer.Customer) mono.Mono[string] {
	f := flow.New[string]()
	var (
		nameMono    mono.Mono[string]
		messageMono mono.Mono[string]
	)

	return f.Steps(func() (await.Group, error) {
		nameMono = concat.Concat(ctx, customer.FirstName, customer.LastName)
		return await.All(nameMono)
	}, func() (await.Group, error) {
		name, _ := nameMono.Get()
		messageMono = greeting.SayHello(ctx, name)
		return await.All(messageMono)
	}, func() (await.Group, error) {
		message, _ := messageMono.Get()
		return f.Success(message)
	}).Mono()
}
