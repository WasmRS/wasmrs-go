package flux

func FromSlice[T any](values []T) Flux[T] {
	return Create(func(sink Sink[T]) {
		sink.OnSubscribe(OnSubscribe{
			Request: func(n int) {
				for i := n; i > 0 && len(values) != 0; i-- {
					next := values[0]
					values = values[1:]
					sink.Next(next)
				}
				if len(values) == 0 {
					sink.Complete()
				}
			},
		})
	})
}
