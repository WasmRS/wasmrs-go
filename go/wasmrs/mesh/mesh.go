package mesh

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
	"github.com/rodaine/table"

	"github.com/nanobus/iota/go/wasmrs/host"
	"github.com/nanobus/iota/go/wasmrs/operations"
	"github.com/nanobus/iota/go/wasmrs/payload"
	"github.com/nanobus/iota/go/wasmrs/rx/flux"
	"github.com/nanobus/iota/go/wasmrs/rx/mono"
)

type (
	Mesh struct {
		verbose     bool
		instances   map[string]*host.Instance
		exports     map[string]map[string]*atomic.Pointer[destination]
		unsatisfied []*pending
	}

	destination struct {
		instance *host.Instance
		index    uint32
	}

	pending struct {
		instance *host.Instance
		oper     operations.Operation
	}
)

type Option func(*Mesh)

func WithVerbose() Option {
	return func(m *Mesh) {
		m.verbose = true
	}
}

func New(opts ...Option) *Mesh {
	m := Mesh{
		instances:   make(map[string]*host.Instance),
		exports:     map[string]map[string]*atomic.Pointer[destination]{},
		unsatisfied: make([]*pending, 0, 10),
	}

	for _, opt := range opts {
		opt(&m)
	}

	return &m
}

func (m *Mesh) RequestResponse(ctx context.Context, namespace, operation string, p payload.Payload) mono.Mono[payload.Payload] {
	ns, ok := m.exports[namespace]
	if !ok {
		return nil
	}

	ptr, ok := ns[operation]
	if !ok {
		return nil
	}
	dest := ptr.Load()

	return dest.RequestResponse(ctx, p)
}

func (m *Mesh) FireAndForget(ctx context.Context, namespace, operation string, p payload.Payload) {
	ns, ok := m.exports[namespace]
	if !ok {
		return
	}

	ptr, ok := ns[operation]
	if !ok {
		return
	}
	dest := ptr.Load()

	dest.FireAndForget(ctx, p)
}

func (m *Mesh) RequestStream(ctx context.Context, namespace, operation string, p payload.Payload) flux.Flux[payload.Payload] {
	ns, ok := m.exports[namespace]
	if !ok {
		return nil
	}

	ptr, ok := ns[operation]
	if !ok {
		return nil
	}
	dest := ptr.Load()

	return dest.RequestStream(ctx, p)
}

func (m *Mesh) RequestChannel(ctx context.Context, namespace, operation string, p payload.Payload, in flux.Flux[payload.Payload]) flux.Flux[payload.Payload] {
	ns, ok := m.exports[namespace]
	if !ok {
		return nil
	}

	ptr, ok := ns[operation]
	if !ok {
		return nil
	}
	dest := ptr.Load()

	return dest.RequestChannel(ctx, p, in)
}

func (m *Mesh) Close() {
	for _, inst := range m.instances {
		inst.Close()
	}
}

func (m *Mesh) LoadModules(ctx context.Context, filenames ...string) error {
	for _, filename := range filenames {
		if _, err := m.LoadModule(ctx, filename); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mesh) LoadModule(ctx context.Context, filename string) (*host.Instance, error) {
	source, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	h, err := host.New(ctx)
	if err != nil {
		return nil, err
	}
	module, err := h.Compile(ctx, source)
	if err != nil {
		return nil, err
	}
	inst, err := module.Instantiate(ctx)
	if err != nil {
		return nil, err
	}

	// Close previously loaded instance.
	existing, ok := m.instances[filename]
	if ok {
		go func(inst *host.Instance) {
			time.Sleep(5 * time.Second)
			inst.Close()
		}(existing)
	}

	m.instances[filename] = inst

	if m.verbose {
		fmt.Println("Loaded " + filename)
	}
	m.linkOperations(inst)

	return inst, nil
}

func (m *Mesh) linkOperations(inst *host.Instance) {
	opers := inst.Operations()
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()

	var tbl table.Table
	if m.verbose {
		tbl = table.New("Namespace", "Operation", "Direction", "Type", "Index")
		tbl.WithHeaderFormatter(headerFmt)
	}

	numExported := 0
	for _, op := range opers {
		if tbl != nil {
			tbl.AddRow(op.Namespace, op.Operation, op.Direction, op.Type, op.Index)
		}

		switch op.Direction {
		case operations.Export:
			ns, ok := m.exports[op.Namespace]
			if !ok {
				ns = make(map[string]*atomic.Pointer[destination])
				m.exports[op.Namespace] = ns
			}
			ptr, ok := ns[op.Operation]
			if !ok {
				ptr = &atomic.Pointer[destination]{}
				ns[op.Operation] = ptr
			}

			ptr.Store(&destination{
				instance: inst,
				index:    op.Index,
			})
			numExported++

		case operations.Import:
			if ok := m.linkOperation(inst, op); !ok {
				m.unsatisfied = append(m.unsatisfied, &pending{
					instance: inst,
					oper:     op,
				})
			}
		}
	}

	if numExported > 0 && len(m.unsatisfied) > 0 {
		filtered := m.unsatisfied[:0]
		for _, u := range m.unsatisfied {
			if ok := m.linkOperation(u.instance, u.oper); !ok {
				filtered = append(filtered, u)
			}
		}
		m.unsatisfied = filtered
	}

	if tbl != nil {
		tbl.Print()
		fmt.Println()
	}
}

func (m *Mesh) linkOperation(inst *host.Instance, op operations.Operation) bool {
	ns, ok := m.exports[op.Namespace]
	if !ok {
		return false
	}

	ptr, ok := ns[op.Operation]
	if !ok {
		return false
	}
	dest := ptr.Load()

	switch op.Type {
	case operations.RequestResponse:
		inst.SetRequestResponseHandler(op.Index, dest.RequestResponse)
	case operations.FireAndForget:
		inst.SetFireAndForgetHandler(op.Index, dest.FireAndForget)
	case operations.RequestStream:
		inst.SetRequestStreamHandler(op.Index, dest.RequestStream)
	case operations.RequestChannel:
		inst.SetRequestChannelHandler(op.Index, dest.RequestChannel)
	}

	return true
}

func (d *destination) RequestResponse(ctx context.Context, p payload.Payload) mono.Mono[payload.Payload] {
	md := p.Metadata()
	if md != nil {
		binary.BigEndian.PutUint32(md, d.index)
	}
	return d.instance.RequestResponse(ctx, p)
}

func (d *destination) FireAndForget(ctx context.Context, p payload.Payload) {
	md := p.Metadata()
	if md != nil {
		binary.BigEndian.PutUint32(md, d.index)
	}
	d.instance.FireAndForget(ctx, p)
}

func (d *destination) RequestStream(ctx context.Context, p payload.Payload) flux.Flux[payload.Payload] {
	md := p.Metadata()
	if md != nil {
		binary.BigEndian.PutUint32(md, d.index)
	}
	return d.instance.RequestStream(ctx, p)
}

func (d *destination) RequestChannel(ctx context.Context, p payload.Payload, in flux.Flux[payload.Payload]) flux.Flux[payload.Payload] {
	md := p.Metadata()
	if md != nil {
		binary.BigEndian.PutUint32(md, d.index)
	}
	return d.instance.RequestChannel(ctx, p, in)
}
