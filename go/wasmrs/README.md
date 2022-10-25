# WebAssembly Reactive Streams (WasmRS) for Go

An implementation of [reactive streams](https://www.reactive-streams.org) implemented in Go WebAssembly modules.

## Protocol details

This implementation differs from the [waPC protocol](https://wapc.io/docs/spec/). Instead of using several host and guest calls to pass request and response buffers, 2 separate circular buffers are used for bidirectional communication (sending data from the host to guest and visa versa). Frames of data are sent over these buffers that are similar to those in other multiplexing protocols like HTTP/2 are passed. In fact, the implementation closely follows the [RSocket protocol](https://rsocket.io/about/protocol).

Frames contain a stream ID, allowing a single Wasm module instance to handle multiple requests on a single thread. Frames also allow for streaming interaction models, which are incredibly useful in many backend processing scenarios (i.e. stream processing on querying databases). The developer uses a Reactive Streams (RS) API that hides the details of the protocol so implementing core logic feels familiar.

### Initialization

The host first calls `__wasmrs_init(guestBufferSize: u32, hostBufferSize: u32): u32` where the arguments are the desired size of the guest and host buffers. The guest allocates the two buffers and calls `__init_buffers(guestBufferPtr: u32, hostBufferPtr: u32)` passing the pointers of the buffers. If successful, `__wasmrs_init` returns `1` or `true` both sides are now ready to communicate by sending frames.

### Sending frames

After initialization, sending frames involves copying data to the circular buffers (starting at offset 0), 1 guest call and 1 host call.

Writing a frame to either the guest and host buffers is accomplished by writing 3 bytes for the length (Big Endian) followed by frame data itself. Being that the buffers are circular, the 3 length bytes or data can wrap around to offset 0 of the buffer. After copying the data, a "send" function is called to signal the host or guest to take action on the new data.

* `(export "__wasmrs_send" (func (param i32))))` - The host calls this exported function passing the offset of the next position it will write to.

* `(import "wasmrs" "__send" (func (param i32))))` - The guest calls this imported function passing the offset of the next position it will write to.

The host and guest implement an event loop reading their circular buffer up to the position passed in the call above. In most cases, frames are small enough to send via the circular buffer with one send call. For larger frames that do not fit into the size of the circular buffer, is called multiple times and the frame data is concatenated internally. Basically, the send functions are repeatedly called whenever the end of the circular buffer is reached and there is remaining frame data to copy. Once the length and frame data are read, they are parsed and handled.

### Interaction models

Frames follow the [RSocket protocol](https://rsocket.io/about/protocol) and allow several requests to be processed concurrently (on a single Wasm thread) as well as concurrent streams of input and output data.

The RSocket protocol supports the following [interaction models](https://rsocket.io/about/motivations#interaction-models):

* **Request/Response** - each request receives a single response.
* **Fire-and-Forget** - the client will receive no response from the server.
* **Request Stream** - a single request may receive multiple responses.
* **Request Channel** - bidirectional streaming communication.

At the application level, the code makes use of reactive streams-style types, namely Mono and Flux. Mono represents the a single payload for Request/Response. Flux represents a stream of responses and returned by Request/Stream and Request Channel or a stream of input to a Request Channel.

## Examples

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
go test -benchmem -run=^$ -bench ^BenchmarkInvoke$ github.com/WasmRS/wasmrs-go/testing -count=1
```
