package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/nanobus/iota/go/operations"
	"github.com/nanobus/iota/go/transport/wasmrs/customsections"
	"github.com/nanobus/iota/go/transport/wasmrs/host"
	"github.com/rodaine/table"
	"github.com/tetratelabs/wabin/binary"
	"github.com/tetratelabs/wabin/wasm"
)

const WASMRS_SECTION_NAME = "wasmrs_ops"

func main() {
	if len(os.Args) < 2 {
		log.Println("usage: wasmrs-ops <wasm file>")
		return
	}
	filename := os.Args[1]

	bin, err := os.ReadFile(filename)
	if err != nil {
		log.Panicln(err)
	}

	sections, err := customsections.Read(bin, WASMRS_SECTION_NAME)
	if err != nil {
		log.Panicln(err)
	}
	for name, section := range sections {
		fmt.Println(name)
		opTable, err := operations.FromBytes(section.Data)
		if err != nil {
			log.Panicln(err)
		}
		printOperationTable(opTable)
	}

	mod, err := binary.DecodeModule(bin, wasm.CoreFeaturesV2)
	if err != nil {
		log.Panicln(err)
	}

	var wasmrs *wasm.CustomSection
	for _, custom := range mod.CustomSections {
		if custom.Name == WASMRS_SECTION_NAME {
			wasmrs = custom
			break
		}
	}

	var opTable operations.Table
	if wasmrs == nil {
		opTable = getOperationListFromModule(context.Background(), bin)
		wasmrs = &wasm.CustomSection{
			Name: WASMRS_SECTION_NAME,
			Data: opTable.ToBytes(),
		}
		mod.CustomSections = append(mod.CustomSections, wasmrs)

		bin = binary.EncodeModule(mod)
		if err = os.WriteFile(filename, bin, 0644); err != nil {
			log.Panicln(err)
		}

		fmt.Printf("Added %s section\n", WASMRS_SECTION_NAME)
	} else {
		if opTable, err = operations.FromBytes(wasmrs.Data); err != nil {
			log.Panicln(err)
		}
	}

	printOperationTable(opTable)
	fmt.Println("")
}

func printOperationTable(opTable operations.Table) {
	fmt.Println("WasmRS operation table")
	fmt.Println("")

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()

	tbl := table.New("Namespace", "Operation", "Direction", "Type", "Index")
	tbl.WithHeaderFormatter(headerFmt)

	for _, op := range opTable {
		tbl.AddRow(op.Namespace, op.Operation, op.Direction, op.Type, op.Index)
	}

	tbl.Print()
}

func getOperationListFromModule(ctx context.Context, bin []byte) operations.Table {
	h, err := host.New(ctx)
	if err != nil {
		log.Panicln(err)
	}
	m, err := h.Compile(ctx, bin)
	if err != nil {
		log.Panicln(err)
	}

	inst, err := m.Instantiate(ctx)
	if err != nil {
		log.Panicln(err)
	}
	return inst.Operations()
}
