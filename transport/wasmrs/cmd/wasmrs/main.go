package main

import (
	"fmt"
	"runtime"

	"github.com/alecthomas/kong"

	"github.com/nanobus/iota/go/transport/wasmrs/cmd/wasmrs/commands"
)

var version = "edge"

var cli struct {
	// List commands.ListCmd `cmd:"" help:"List info contained in a WasmRS module."`
	// Invoke reinstalls the base module dependencies.
	Invoke commands.InvokeCmd `cmd:"" help:"Invokes a WasmRS module."`
	// Version prints out the version of this program and runtime info.
	Version versionCmd `cmd:""`
}

func main() {
	ctx := kong.Parse(&cli)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run(&commands.Context{})
	ctx.FatalIfErrorf(err)
}

type versionCmd struct{}

func (c *versionCmd) Run() error {
	fmt.Printf("wasmrs version %s %s/%s\n", version, runtime.GOOS, runtime.GOARCH)
	return nil
}
