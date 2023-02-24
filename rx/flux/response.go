package flux

func SliceResponse[T any](sink Sink[T], slice []T) func(n int) {
	return func(n int) {
		num := len(slice)
		if n < num {
			num = n
		}
		for i := 0; i < num; i++ {
			sink.Next(slice[i])
		}
		slice = slice[num:]
		if len(slice) == 0 {
			sink.Complete()
		}
	}
}
