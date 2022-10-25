package host

import (
	"context"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

const (
	functionStart = "_start"
	functionInit  = "wasmrs_init"
)

type Host struct {
	runtime wazero.Runtime
	config  wazero.ModuleConfig
}

type Module struct {
	h      *Host
	module wazero.CompiledModule
}

func New(ctx context.Context) (*Host, error) {
	rc := wazero.NewRuntimeConfig().WithCoreFeatures(api.CoreFeaturesV2)
	r := wazero.NewRuntimeWithConfig(ctx, rc)
	// Call any WASI or WasmRS start functions on instantiate.
	config := wazero.NewModuleConfig().
		WithStartFunctions(functionStart, functionInit).
		WithStdin(os.Stdin).
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		WithSysWalltime().
		WithSysNanotime()

	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		return nil, err
	}

	_, err := wasmrs(r.NewHostModuleBuilder("wasmrs")).
		Instantiate(ctx, r)
	if err != nil {
		return nil, err
	}

	return &Host{
		runtime: r,
		config:  config,
	}, nil
}

func (h *Host) Compile(ctx context.Context, source []byte) (*Module, error) {
	compiled, err := h.runtime.CompileModule(ctx, source)
	if err != nil {
		return nil, err
	}

	return &Module{
		h:      h,
		module: compiled,
	}, nil
}

func (m *Module) Instantiate(ctx context.Context) (*Instance, error) {
	module, err := m.h.runtime.InstantiateModule(ctx, m.module, m.h.config.WithName("test"))
	if err != nil {
		return nil, err
	}

	return NewInstance(ctx, module)
}

func wasmrs(builder wazero.HostModuleBuilder) wazero.HostModuleBuilder {
	return builder.
		ExportFunction("__init_buffers", func(ctx context.Context, m api.Module, sendPtr, recvPtr uint32) {
			i := ctx.Value(instanceKey{}).(*Instance)
			i.setBuffers(sendPtr, recvPtr)
		}).
		ExportFunction("__op_list", func(ctx context.Context, m api.Module, opPtr uint32, opSize uint32) {
			i := ctx.Value(instanceKey{}).(*Instance)
			i.opList(ctx, opPtr, opSize)
		}).
		ExportFunction("__send", func(ctx context.Context, m api.Module, recvPos uint32) {
			i := ctx.Value(instanceKey{}).(*Instance)
			i.hostSend(ctx, recvPos)
		})
}
