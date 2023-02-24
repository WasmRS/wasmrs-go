package flux

type Processor[T any] interface {
	Flux[T]
	Sink[T]
}

type processor[T any] struct {
	Flux[T]
	sink Sink[T]
	sub  *OnSubscribe
}

func NewProcessor[T any]() Processor[T] {
	p := &processor[T]{}
	p.Flux = Create(func(sink Sink[T]) {
		p.sink = sink
		if p.sub != nil {
			sink.OnSubscribe(*p.sub)
			p.sub = nil
		}
	})

	return p
}

func (p *processor[T]) Next(value T) {
	if p.sink != nil {
		p.sink.Next(value)
	}
}

func (p *processor[T]) Complete() {
	if p.sink != nil {
		p.sink.Complete()
	}
}

func (p *processor[T]) Error(err error) {
	if p.sink != nil {
		p.sink.Error(err)
	}
}

func (p *processor[T]) OnSubscribe(sub OnSubscribe) {
	if p.sink != nil {
		p.sink.OnSubscribe(sub)
	} else {
		p.sub = &sub
	}
}
