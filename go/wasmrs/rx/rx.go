package rx

import (
	"errors"
	"math"
)

const RequestMax = math.MaxInt32

var ErrTimeout = errors.New("timeout")

type SignalType int

const (
	// SignalComplete indicated that subscriber was completed.
	SignalComplete SignalType = iota
	// SignalCancel indicates that subscriber was cancelled.
	SignalCancel
	// SignalError indicates that subscriber has some faults.
	SignalError
)

type (
	// FnOnComplete is alias of function for signal when no more elements are available
	FnOnComplete func()
	// FnOnNext is alias of function for signal when next element arrived.
	FnOnNext[T any] func(T)
	// FnOnSubscribe is alias of function for signal when subscribe begin.
	FnOnSubscribe func(s Subscription)
	//FnOnSubscribe func(ctx context.Context, s Subscription)
	// FnOnError is alias of function for signal when an error occurred.
	FnOnError func(error)
	// FnOnCancel is alias of function for signal when subscription canceled.
	FnOnCancel func(FnCancel)
	// FnFinally is alias of function for signal when all things done.
	FnFinally   func(SignalType)
	FnOnRequest func(cancel func())

	FnCancel func()
)

// Subscription represents a one-to-one lifecycle of a Subscriber subscribing to a Publisher.
type Subscription interface {
	// Request requests the next N items.
	Request(n int)
	// Cancel cancels the current lifecycle of subscribing.
	Cancel()
}

type HasSubscription interface {
	Subscription() Subscription
}

type Transform[S any, D any] func(S) (D, error)

func (s SignalType) String() string {
	switch s {
	case SignalComplete:
		return "COMPLETE"
	case SignalCancel:
		return "CANCEL"
	case SignalError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}
