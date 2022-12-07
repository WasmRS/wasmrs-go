package await

import (
	"github.com/nanobus/iota/go/rx"
)

type NotifyCallback func()

type Awaitable interface {
	Async()
	Notify(rx.FnFinally)
}

type Group []Awaitable

func All(awaitables ...Awaitable) (Group, error) {
	return awaitables, nil
}
