#!/bin/bash
#rm cmd/greeter/main.wasm

tinygo build -o cmd/greeter/main.wasm -scheduler=none -target wasi -llvm-features "+bulk-memory" -no-debug cmd/greeter/main.go
go test -benchmem -run=^$ -bench ^BenchmarkInvoke$ github.com/WasmRS/wasmrs-go/testing -count=1