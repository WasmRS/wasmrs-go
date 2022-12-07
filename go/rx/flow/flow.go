package flow

import (
	"github.com/nanobus/iota/go/rx"
	"github.com/nanobus/iota/go/rx/await"
	"github.com/nanobus/iota/go/rx/flux"
	"github.com/nanobus/iota/go/rx/mono"
)

type (
	Flow[T any] struct {
		steps    []StepFn
		monoSink mono.Sink[T]
		fluxSink flux.Sink[T]

		ended      bool
		completed  int
		awaitables []await.Awaitable
	}

	StepFn func() (await.Group, error)
)

func (f *Flow[T]) End() (await.Group, error) {
	f.ended = true
	return nil, nil
}

func New[T any]() *Flow[T] {
	return &Flow[T]{}
}

func (f *Flow[T]) Steps(steps ...StepFn) *Flow[T] {
	f.steps = append(f.steps, steps...)
	return f
}

func (f *Flow[T]) Next(value T) {
	if f.fluxSink == nil {
		panic("Next is only called for flux retruns")
	}
	f.fluxSink.Next(value)
}

func (f *Flow[T]) Complete() (await.Group, error) {
	if f.fluxSink == nil {
		panic("Complete is only called for flux retruns")
	}
	f.fluxSink.Complete()

	return nil, nil
}

func (f *Flow[T]) Success(value T) (await.Group, error) {
	if f.monoSink == nil {
		panic("success is only called for mono retruns")
	}
	f.monoSink.Success(value)

	return nil, nil
}

func (f *Flow[T]) Error(err error) (await.Group, error) {
	if f.monoSink != nil {
		f.monoSink.Error(err)
	} else if f.fluxSink != nil {
		f.fluxSink.Error(err)
	}

	return nil, err
}

func (f *Flow[T]) Mono() mono.Mono[T] {
	return mono.Create(func(s mono.Sink[T]) {
		f.monoSink = s
		f.next()
	})
}

func (f *Flow[T]) Flux() flux.Flux[T] {
	return flux.Create(func(s flux.Sink[T]) {
		f.fluxSink = s
		f.next()
	})
}

func (f *Flow[T]) next() {
	if len(f.steps) == 0 {
		return
	}

	f.completed = 0
	step := f.steps[0]
	awaitables, err := step()
	if err != nil {
		if f.monoSink != nil {
			f.monoSink.Error(err)
		} else if f.fluxSink != nil {
			f.fluxSink.Error(err)
		}
		return
	}

	if f.ended {
		return
	}

	f.awaitables = awaitables
	for _, a := range f.awaitables {
		a.Async() // Ensure task is running.
		a.Notify(f.check)
	}

	f.steps = f.steps[1:]
}

func (f *Flow[T]) check(rx.SignalType) {
	f.completed++
	if f.completed >= len(f.awaitables) {
		f.next()
	}
}
