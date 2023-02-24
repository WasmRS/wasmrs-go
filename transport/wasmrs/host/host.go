package host

import (
	"context"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/assemblyscript"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

const (
	i32 = api.ValueTypeI32

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

	// This disables the abort message as no other engines write it.
	envBuilder := r.NewHostModuleBuilder("env")
	assemblyscript.NewFunctionExporter().WithAbortMessageDisabled().ExportFunctions(envBuilder)
	if _, err := envBuilder.Instantiate(ctx); err != nil {
		_ = r.Close(ctx)
		return nil, err
	}

	if _, err := instantiateWasmrs(ctx, r); err != nil {
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

func instantiateWasmrs(ctx context.Context, r wazero.Runtime) (api.Closer, error) {
	return r.NewHostModuleBuilder("wasmrs").
		NewFunctionBuilder().
		WithGoFunction(api.GoFunc(initBuffers), []api.ValueType{i32, i32}, []api.ValueType{}).
		WithParameterNames("send_ptr", "recv_ptr").Export("__init_buffers").
		NewFunctionBuilder().
		WithGoFunction(api.GoFunc(opList), []api.ValueType{i32, i32}, []api.ValueType{}).
		WithParameterNames("op_ptr", "op_size").Export("__op_list").
		NewFunctionBuilder().
		WithGoFunction(api.GoFunc(send), []api.ValueType{i32}, []api.ValueType{}).
		WithParameterNames("recv_pos").Export("__send").
		Instantiate(ctx)
}

// initBuffers is defined as an api.GoFunc for better performance vs reflection.
func initBuffers(ctx context.Context, params []uint64) {
	sendPtr, recvPtr := uint32(params[0]), uint32(params[1])

	i := ctx.Value(instanceKey{}).(*Instance)
	i.setBuffers(sendPtr, recvPtr)
}

// opList is defined as an api.GoFunc for better performance vs reflection.
func opList(ctx context.Context, params []uint64) {
	opPtr, opSize := uint32(params[0]), uint32(params[1])

	i := ctx.Value(instanceKey{}).(*Instance)
	i.opList(ctx, opPtr, opSize)
}

// send is defined as an api.GoFunc for better performance vs reflection.
func send(ctx context.Context, params []uint64) {
	recvPos := uint32(params[0])

	i := ctx.Value(instanceKey{}).(*Instance)
	i.hostSend(ctx, recvPos)
}
