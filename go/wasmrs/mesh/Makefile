.PHONY: build concat

build: concat greeting customer

concat:
	tinygo build -o concat.wasm -scheduler=none -target wasi -no-debug modules/concat/cmd/main.go
	wasm-opt -O concat.wasm -o concat.wasm
	go run ../cmd/wasmrs-ops/main.go concat.wasm

greeting:
	tinygo build -o greeting.wasm -scheduler=none -target wasi -no-debug modules/greeting/cmd/main.go
	wasm-opt -O greeting.wasm -o greeting.wasm
	go run ../cmd/wasmrs-ops/main.go greeting.wasm

customer:
	tinygo build -o customer.wasm -scheduler=none -target wasi -no-debug modules/customer/cmd/main.go
	wasm-opt -O customer.wasm -o customer.wasm
	go run ../cmd/wasmrs-ops/main.go customer.wasm
