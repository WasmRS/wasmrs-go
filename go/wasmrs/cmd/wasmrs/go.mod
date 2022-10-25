module github.com/WasmRS/wasmrs-go/cmd/wasmrs

go 1.19

replace github.com/WasmRS/wasmrs-go => ../../

require (
	github.com/WasmRS/wasmrs-go v0.0.0-00010101000000-000000000000
	github.com/alecthomas/kong v0.6.1
	github.com/vmihailenco/msgpack/v5 v5.3.5
	github.com/nanobus/iota/go/msgpack v0.1.3
)

require (
	github.com/fatih/color v1.13.0 // indirect
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/rodaine/table v1.0.1 // indirect
	github.com/tetratelabs/wazero v1.0.0-pre.2 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	golang.org/x/exp v0.0.0-20220722155223-a9213eeb770e // indirect
	golang.org/x/sys v0.0.0-20211019181941-9d821ace8654 // indirect
)
