package main

import (
	"context"
	"strings"

	"github.com/nanobus/iota/go/rx"
	"github.com/nanobus/iota/go/rx/flux"
	"github.com/nanobus/iota/go/rx/mono"
	"github.com/nanobus/iota/go/transport/wasmrs/await"
	"github.com/nanobus/iota/go/transport/wasmrs/flow"
	"github.com/nanobus/iota/go/transport/wasmrs/guest"

	"github.com/nanobus/iota/go/transport/wasmrs/example"
)

var (
	counter *example.Counter   = example.NewCounter(guest.HostInvoker)
	ucase   *example.Uppercase = example.NewUppercase(guest.HostInvoker)
)

func main() {
	example.RegisterSayHello(SayHello)
	example.RegisterUppercase(Uppercase)
}

// This example uses a process utility helper that uses closures to execute
// groups of functionality and return when an await is needed. The utility
// handles all the callback work and invokes the next step when the required
// async calls are completed.
func SayHello(ctx context.Context, firstName, lastName string) mono.Mono[string] {
	f := flow.New[string]()
	var (
		test      int
		uppercase flux.Flux[string]
		counter0  flux.Flux[string]
		counter1  flux.Flux[string]
		counter2  flux.Flux[string]
	)

	return f.Steps(func() (await.Group, error) {
		println("uppercase")
		in := flux.FromSlice([]string{"one", "two", "three"})
		uppercase = ucase.Uppercase(ctx, in).Subscribe(flux.Subscribe[string]{
			OnNext: func(value string) {
				println(value)
			},
			OnRequest: func(sub rx.Subscription) {
				sub.Request(2)
			},
		})

		println("counter 0")
		// Async counter 1: 1...3
		counter0 = counter.CountToN(ctx, 3).Subscribe(flux.Subscribe[string]{
			OnNext: func(value string) {
				test++
				println("Counter 0", value)
			},
			OnRequest: func(s rx.Subscription) {
				s.Request(2)
			},
		})

		println("counter 1")
		// Async counter 1: 1...3
		counter1 = counter.CountToN(ctx, 3).Subscribe(flux.Subscribe[string]{
			OnNext: func(value string) {
				test++
				println("Counter 1", value)
			},
			OnRequest: func(s rx.Subscription) {
				s.Request(2)
			},
		})

		println("counter 2")
		// Async counter 2: 1...5
		counter2 = counter.CountToN(ctx, 5).Subscribe(flux.Subscribe[string]{
			OnNext: func(value string) {
				test++
				println("Counter 2", value)
			},
			OnRequest: func(s rx.Subscription) {
				s.Request(2)
			},
		})

		return await.All(uppercase, counter0, counter1, counter2)
	}, func() (await.Group, error) {
		return f.Success("Hello, " + firstName + " " + lastName)
	}).Mono()
}

func Uppercase(ctx context.Context, in flux.Flux[string]) flux.Flux[string] {
	return flux.Create(func(sink flux.Sink[string]) {
		in.Subscribe(flux.Subscribe[string]{
			OnNext: func(value string) {
				sink.Next(strings.ToUpper(value))
			},
			OnComplete: func() {
				println("Complete")
				sink.Complete()
			},
			OnError: func(err error) {
				sink.Error(err)
			},
			OnRequest: func(s rx.Subscription) {
				s.Request(2)
			},
		})
	})
}
