# WebAssembly Reactive Streams (WasmRS) for Go

An implementation of [reactive streams](https://www.reactive-streams.org) implemented in Go WebAssembly modules.

## Protocol details

See the [WasmRS Protocol Documentation](https://github.com/nanobus/iota/blob/main/docs/wasmrs.md) for details.

## Usage Examples

Simple Request/Response

```go
func SayHello(ctx context.Context, firstName, lastName string) mono.Mono[string] {
	greeting := "Hello, " + firstName + " " + lastName
	return mono.Just(greeting)
}
```

Async Request/Response

```go
func RegisterUser(ctx context.Context, firstName, lastName string) mono.Mono[string] {
	// Awaitables
	var username mono.Mono[string]

	f := flow.New[string]()
	return f.Steps(func() (await.Group, error) {
		username = account.New(firstName, lastName)
		return await.All(username)
	}, func() (await.Group, error) {
		un, err := username.Get()
		if err != nil {
			return f.Error(err)
		}

		greeting := "Hello, " + firstName + " " + lastName + ". " +
			"Your username is " + un
		return f.Success(greeting)
	}).Mono()
}
```

Request Stream

```go
func CountToN(ctx context.Context, to uint64) flux.Flux[string] {
	return flux.Create(func(sink flux.Sink[string]) {
		i := uint64(1)
		sink.OnSubscribe(flux.OnSubscribe{
			Request: func(n int) {
				for ; i <= to && n > 0; i++ {
					msg := "Num: " + i
					sink.Next(msg)
					n--
				}
				if i > to {
					sink.Complete()
				}
			},
		})
	})
}
```

Request Channel

```go
func Uppercase(ctx context.Context, in flux.Flux[string]) flux.Flux[string] {
	return flux.Create(func(sink flux.Sink[string]) {
		in.Subscribe(flux.Subscribe[string]{
			OnNext: func(value string) {
				sink.Next(strings.ToUpper(value))
			},
			OnComplete: func() {
				sink.Complete()
			},
			OnError: func(err error) {
				sink.Error(err)
			},
			OnRequest: func(s rx.Subscription) {
				s.Request(10)
			},
		})
	})
}
```

## Running

Verify you have the TinyGo 0.26.0 and go 1.19 or later.

```shell
> tinygo version
tinygo version 0.26.0 darwin/amd64 (using go version go1.18.5 and LLVM version 14.0.6)
```

You can follow [these instructions](https://tinygo.org/getting-started/install/) to install TinyGo.

### Building the wasm module

```shell
tinygo build -o cmd/guest/main.wasm -scheduler=none -target wasi -no-debug cmd/guest/main.go
```

### Running the host

```shell
go run cmd/host/main.go
```

### FBP example

```shell
tinygo build -o cmd/fbp/guest/main.wasm -scheduler=none -target wasi -no-debug cmd/fbp/guest/main.go
go run cmd/fbp/host/main.go
```

### Bench test

```
tinygo build -o cmd/greeter/main.wasm -scheduler=none -target wasi -no-debug cmd/greeter/main.go
go test -benchmem -run=^$ -bench ^BenchmarkInvoke$ github.com/nanobus/iota/go/wasmrs/testing -count=1
```
