package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/nanobus/iota/go/invoke"
	"github.com/nanobus/iota/go/transport/rsocket"

	"github.com/nanobus/iota/go/example/greeter"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	server := rsocket.NewServer(func(ctx context.Context, caller invoke.Caller, onClose func(*rsocket.Transport)) {
		fmt.Println("Got connection")
		client := greeter.NewGreeting(caller)

		msg, err := client.SayHello(ctx, "RSocket", "Server").Block()
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(msg)
	})

	notifier := make(chan bool, 1)
	go server.Listen(ctx, notifier)
	listening := <-notifier
	if !listening {
		panic("could not open listening socket")
	}

	fmt.Println("waiting for connections")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
