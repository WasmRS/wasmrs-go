package mono

type Processor[T any] interface {
	Mono[T]
	Sink[T]
}

type ProcessorImpl[T any] struct {
	Mono[T]
	sink    Sink[T]
	success *T
	err     error
}

func NewProcessor[T any]() *ProcessorImpl[T] {
	p := &ProcessorImpl[T]{}
	p.Mono = Create(func(sink Sink[T]) {
		p.sink = sink
		if p.success != nil {
			sink.Success(*p.success)
		} else if p.err != nil {
			sink.Error(p.err)
		}
	})

	return p
}

func (p *ProcessorImpl[T]) Success(value T) {
	if p.sink != nil {
		p.sink.Success(value)
	} else {
		p.success = &value
	}
}

func (p *ProcessorImpl[T]) Error(err error) {
	if p.sink != nil {
		p.sink.Error(err)
	} else {
		p.err = err
	}
}
