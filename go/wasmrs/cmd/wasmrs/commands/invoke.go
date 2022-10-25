package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/nanobus/iota/go/wasmrs/mesh"
	"github.com/nanobus/iota/go/wasmrs/payload"
)

type Context struct{}

type InvokeCmd struct {
	Pretty    bool     `help:"Pretty print the output."`
	Verbose   bool     `help:"Print verbose output."`
	Namespace string   `arg:"" help:"The namespace of the operation to invoke"`
	Operation string   `arg:"" help:"The name of the operation to invoke"`
	Modules   []string `arg:"" type:"existingfile" help:"The WasmRS modules to load"`
}

func (c *InvokeCmd) Run() error {
	ctx := context.Background()
	opts := []mesh.Option{}
	if c.Verbose {
		opts = append(opts, mesh.WithVerbose())
	}
	m := mesh.New(opts...)
	defer m.Close()

	// Modules can be loaded in any order
	// so long as all imports are satisfied.
	if err := m.LoadModules(ctx, c.Modules...); err != nil {
		return err
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("could not read stdin: %w", err)
	}
	var dataIface interface{}
	if err = json.Unmarshal(data, &dataIface); err != nil {
		return fmt.Errorf("could not parse stdin: %w", err)
	}
	input, err := msgpack.Marshal(dataIface)
	if err != nil {
		return fmt.Errorf("could not convert input to MsgPack: %w", err)
	}

	var metadata [8]byte
	p := payload.New(input, metadata[:])

	result, err := m.RequestResponse(ctx, c.Namespace, c.Operation, p).Block()
	if err != nil {
		log.Panic(err)
	}

	var outputIface interface{}
	if err = msgpack.Unmarshal(result.Data(), &outputIface); err != nil {
		return fmt.Errorf("could not parse module output: %w", err)
	}

	var jsonBytes []byte
	if c.Pretty {
		jsonBytes, err = json.MarshalIndent(outputIface, "", "  ")
	} else {
		jsonBytes, err = json.Marshal(outputIface)
	}
	if err != nil {
		return fmt.Errorf("error converting output to JSON: %w", err)
	}

	fmt.Println(string(jsonBytes))

	return nil
}
