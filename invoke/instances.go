package invoke

import (
	"errors"
	"sync"
	"unsafe"
)

type Ptr[T any] struct {
	Ptr T
}

var ErrNoInstance = errors.New("no instance is active for this pointer")

type LiveInstances[T any] struct {
	mu        sync.RWMutex
	instances map[*Ptr[T]]struct{}
}

func NewLiveInstances[T any]() *LiveInstances[T] {
	return &LiveInstances[T]{
		instances: make(map[*Ptr[T]]struct{}),
	}
}

func (t *LiveInstances[T]) Get(id uint64) (inst T, ok bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	ptr := (*Ptr[T])(unsafe.Pointer(uintptr(id)))
	_, ok = t.instances[ptr]
	if ok {
		inst = ptr.Ptr
	}
	return
}

func (t *LiveInstances[T]) Put(inst T) uint64 {
	ptr := &Ptr[T]{Ptr: inst}
	handle := uint64(uintptr(unsafe.Pointer(ptr)))
	t.mu.Lock()
	defer t.mu.Unlock()
	t.instances[ptr] = struct{}{}
	return handle
}

func (t *LiveInstances[T]) Remove(id uint64) (inst T, ok bool) {
	ptr := (*Ptr[T])(unsafe.Pointer(uintptr(id)))
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok = t.instances[ptr]; ok {
		inst = ptr.Ptr
		delete(t.instances, ptr)
	}
	return
}
