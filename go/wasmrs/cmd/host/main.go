package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nanobus/iota/go/wasmrs/host"
	"github.com/nanobus/iota/go/wasmrs/rx"
	"github.com/nanobus/iota/go/wasmrs/rx/flux"

	"github.com/nanobus/iota/go/wasmrs/example"
)

func main() {
	source, err := os.ReadFile("cmd/guest/main.wasm")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()

	h, err := host.New(ctx)
	if err != nil {
		panic(err)
	}
	module, err := h.Compile(ctx, source)
	if err != nil {
		panic(err)
	}
	inst, err := module.Instantiate(ctx)
	if err != nil {
		panic(err)
	}

	g := example.NewGreeting(inst)

	// Register host calls
	example.RegisterCounter(CountToN)
	example.RegisterUppercase(Uppercase)

	// Call guest module
	message, err := g.SayHello(ctx, "John", "Doe").Block()
	if err == nil {
		log.Println("Success:", message)
	} else {
		log.Println("Error:", err.Error())
	}

	in := flux.FromSlice([]string{"one", "two", "three"})
	g.Uppercase(ctx, in).Block(flux.Subscribe[string]{
		OnNext: func(value string) {
			fmt.Println(value)
		},
		OnRequest: func(sub rx.Subscription) {
			sub.Request(2)
		},
	})
}

func CountToN(ctx context.Context, to uint64) flux.Flux[string] {
	return flux.Create(func(sink flux.Sink[string]) {
		i := uint64(1)
		sink.OnSubscribe(flux.OnSubscribe{
			Request: func(n int) {
				for ; i <= to && n > 0; i++ {
					t := time.Now().Format("2006/01/02 15:04:05")
					msg := fmt.Sprintf("%s Num: %d", t, i)
					sink.Next(msg)
					time.Sleep(time.Second)
					n--
				}
				if i > to {
					fmt.Println("Complete")
					sink.Complete()
				}
			},
		})
	})
}

func Uppercase(ctx context.Context, in flux.Flux[string]) flux.Flux[string] {
	return flux.Create(func(sink flux.Sink[string]) {
		in.Subscribe(flux.Subscribe[string]{
			OnNext: func(value string) {
				sink.Next(strings.ToUpper(value))
			},
			OnComplete: func() {
				fmt.Println("Complete")
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
