package main

import (
	"fmt"
	"time"

	"github.com/nanobus/iota/go/rx"
	"github.com/nanobus/iota/go/rx/flux"
	"github.com/nanobus/iota/go/rx/mono"
)

func main() {
	flow := flux.Create(func(subscriber flux.Sink[int]) {
		values := []int{0, 1, 2, 3, 4, 5}
		subscriber.OnSubscribe(flux.OnSubscribe{
			Cancel: func() {},
			Request: func(n int) {
				fmt.Println("Remaining", values)
				for i := n; i > 0 && len(values) != 0; i-- {
					next := values[0]
					values = values[1:]
					subscriber.Next(next)
					time.Sleep(time.Second)
				}
				if len(values) == 0 {
					subscriber.Complete()
				}
			},
		})
	})

	const pageSize = 2
	done := make(chan struct{})
	flux.Map(flow, func(i int) (int, error) {
		return i + 100000, nil
	}).Subscribe(flux.Subscribe[int]{
		OnComplete: func() { fmt.Println("done") },
		OnNext:     func(value int) { fmt.Println("value =", value) },
		OnError:    func(err error) { fmt.Println("Error:", err) },
		// Nothing happens until `Request(n)` is called
		OnRequest: func(sub rx.Subscription) {
			sub.Request(pageSize)
		},
		Finally: func(signal rx.SignalType) {
			fmt.Println("signal =", signal)
			close(done)
		},
	})

	<-done
	fmt.Println("After")

	fmt.Println("-----------------------------")

	sing := mono.Create(func(subscriber mono.Sink[int]) {
		subscriber.Success(1234)
	})

	mono.Map(sing, func(i int) (int, error) {
		return i + 100000, nil
	}).Subscribe(mono.Subscribe[int]{
		OnSuccess: func(value int) { fmt.Println("value =", value) },
		OnError:   func(err error) { fmt.Println("Error:", err) },
		OnRequest: func(cancel rx.FnCancel) {},
		Finally: func(signal rx.SignalType) {
			fmt.Println("signal =", signal)
		},
	})
}
