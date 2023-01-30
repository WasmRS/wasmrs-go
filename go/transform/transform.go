package transform

import (
	"time"

	"github.com/nanobus/iota/go/invoke"
	"github.com/nanobus/iota/go/msgpack"
	"github.com/nanobus/iota/go/payload"
	"github.com/nanobus/iota/go/rx"
	"github.com/nanobus/iota/go/rx/flux"
	"github.com/nanobus/iota/go/rx/mono"
)

type Transform[T any] struct {
	Decode func(raw payload.Payload) (T, error)
	Encode func(decoded T) (payload.Payload, error)
}

var String = Transform[string]{
	Decode: func(raw payload.Payload) (string, error) {
		decoder := msgpack.NewDecoder(raw.Data())
		val, err := decoder.ReadString()
		return val, err
	},
	Encode: func(value string) (payload.Payload, error) {
		sizer := msgpack.NewSizer()
		sizer.WriteString(value)
		buf := make([]byte, sizer.Len())
		encoder := msgpack.NewEncoder(buf)
		encoder.WriteString(value)
		return payload.New(buf), nil
	},
}

var Int = Transform[int]{
	Decode: func(raw payload.Payload) (int, error) {
		decoder := msgpack.NewDecoder(raw.Data())
		val, err := decoder.ReadInt64()
		return int(val), err
	},
	Encode: func(value int) (payload.Payload, error) {
		sizer := msgpack.NewSizer()
		sizer.WriteInt64(int64(value))
		buf := make([]byte, sizer.Len())
		encoder := msgpack.NewEncoder(buf)
		encoder.WriteInt64(int64(value))
		return payload.New(buf), nil
	},
}

var Int64 = Transform[int64]{
	Decode: func(raw payload.Payload) (int64, error) {
		decoder := msgpack.NewDecoder(raw.Data())
		val, err := decoder.ReadInt64()
		return val, err
	},
	Encode: func(value int64) (payload.Payload, error) {
		sizer := msgpack.NewSizer()
		sizer.WriteInt64(value)
		buf := make([]byte, sizer.Len())
		encoder := msgpack.NewEncoder(buf)
		encoder.WriteInt64(value)
		return payload.New(buf), nil
	},
}

var Int32 = Transform[int32]{
	Decode: func(raw payload.Payload) (int32, error) {
		decoder := msgpack.NewDecoder(raw.Data())
		val, err := decoder.ReadInt32()
		return val, err
	},
	Encode: func(value int32) (payload.Payload, error) {
		sizer := msgpack.NewSizer()
		sizer.WriteInt32(value)
		buf := make([]byte, sizer.Len())
		encoder := msgpack.NewEncoder(buf)
		encoder.WriteInt32(value)
		return payload.New(buf), nil
	},
}

var Int16 = Transform[int16]{
	Decode: func(raw payload.Payload) (int16, error) {
		decoder := msgpack.NewDecoder(raw.Data())
		val, err := decoder.ReadInt16()
		return val, err
	},
	Encode: func(value int16) (payload.Payload, error) {
		sizer := msgpack.NewSizer()
		sizer.WriteInt16(value)
		buf := make([]byte, sizer.Len())
		encoder := msgpack.NewEncoder(buf)
		encoder.WriteInt16(value)
		return payload.New(buf), nil
	},
}

var Int8 = Transform[int8]{
	Decode: func(raw payload.Payload) (int8, error) {
		decoder := msgpack.NewDecoder(raw.Data())
		val, err := decoder.ReadInt8()
		return val, err
	},
	Encode: func(value int8) (payload.Payload, error) {
		sizer := msgpack.NewSizer()
		sizer.WriteInt8(value)
		buf := make([]byte, sizer.Len())
		encoder := msgpack.NewEncoder(buf)
		encoder.WriteInt8(value)
		return payload.New(buf), nil
	},
}

var Uint = Transform[uint]{
	Decode: func(raw payload.Payload) (uint, error) {
		decoder := msgpack.NewDecoder(raw.Data())
		val, err := decoder.ReadUint64()
		return uint(val), err
	},
	Encode: func(value uint) (payload.Payload, error) {
		sizer := msgpack.NewSizer()
		sizer.WriteUint64(uint64(value))
		buf := make([]byte, sizer.Len())
		encoder := msgpack.NewEncoder(buf)
		encoder.WriteUint64(uint64(value))
		return payload.New(buf), nil
	},
}

var Uint64 = Transform[uint64]{
	Decode: func(raw payload.Payload) (uint64, error) {
		decoder := msgpack.NewDecoder(raw.Data())
		val, err := decoder.ReadUint64()
		return val, err
	},
	Encode: func(value uint64) (payload.Payload, error) {
		sizer := msgpack.NewSizer()
		sizer.WriteUint64(value)
		buf := make([]byte, sizer.Len())
		encoder := msgpack.NewEncoder(buf)
		encoder.WriteUint64(value)
		return payload.New(buf), nil
	},
}

var Uint32 = Transform[uint32]{
	Decode: func(raw payload.Payload) (uint32, error) {
		decoder := msgpack.NewDecoder(raw.Data())
		val, err := decoder.ReadUint32()
		return val, err
	},
	Encode: func(value uint32) (payload.Payload, error) {
		sizer := msgpack.NewSizer()
		sizer.WriteUint32(value)
		buf := make([]byte, sizer.Len())
		encoder := msgpack.NewEncoder(buf)
		encoder.WriteUint32(value)
		return payload.New(buf), nil
	},
}

var Uint16 = Transform[uint16]{
	Decode: func(raw payload.Payload) (uint16, error) {
		decoder := msgpack.NewDecoder(raw.Data())
		val, err := decoder.ReadUint16()
		return val, err
	},
	Encode: func(value uint16) (payload.Payload, error) {
		sizer := msgpack.NewSizer()
		sizer.WriteUint16(value)
		buf := make([]byte, sizer.Len())
		encoder := msgpack.NewEncoder(buf)
		encoder.WriteUint16(value)
		return payload.New(buf), nil
	},
}

var Uint8 = Transform[uint8]{
	Decode: func(raw payload.Payload) (uint8, error) {
		decoder := msgpack.NewDecoder(raw.Data())
		val, err := decoder.ReadUint8()
		return val, err
	},
	Encode: func(value uint8) (payload.Payload, error) {
		sizer := msgpack.NewSizer()
		sizer.WriteUint8(value)
		buf := make([]byte, sizer.Len())
		encoder := msgpack.NewEncoder(buf)
		encoder.WriteUint8(value)
		return payload.New(buf), nil
	},
}

var Float64 = Transform[float64]{
	Decode: func(raw payload.Payload) (float64, error) {
		decoder := msgpack.NewDecoder(raw.Data())
		val, err := decoder.ReadFloat64()
		return val, err
	},
	Encode: func(value float64) (payload.Payload, error) {
		sizer := msgpack.NewSizer()
		sizer.WriteFloat64(value)
		buf := make([]byte, sizer.Len())
		encoder := msgpack.NewEncoder(buf)
		encoder.WriteFloat64(value)
		return payload.New(buf), nil
	},
}

var Float32 = Transform[float32]{
	Decode: func(raw payload.Payload) (float32, error) {
		decoder := msgpack.NewDecoder(raw.Data())
		val, err := decoder.ReadFloat32()
		return val, err
	},
	Encode: func(value float32) (payload.Payload, error) {
		sizer := msgpack.NewSizer()
		sizer.WriteFloat32(value)
		buf := make([]byte, sizer.Len())
		encoder := msgpack.NewEncoder(buf)
		encoder.WriteFloat32(value)
		return payload.New(buf), nil
	},
}

var Bytes = Transform[[]byte]{
	Decode: func(raw payload.Payload) ([]byte, error) {
		decoder := msgpack.NewDecoder(raw.Data())
		val, err := decoder.ReadByteArray()
		return val, err
	},
	Encode: func(value []byte) (payload.Payload, error) {
		sizer := msgpack.NewSizer()
		sizer.WriteByteArray(value)
		buf := make([]byte, sizer.Len())
		encoder := msgpack.NewEncoder(buf)
		encoder.WriteByteArray(value)
		return payload.New(buf), nil
	},
}

var Bool = Transform[bool]{
	Decode: func(raw payload.Payload) (bool, error) {
		decoder := msgpack.NewDecoder(raw.Data())
		val, err := decoder.ReadBool()
		return val, err
	},
	Encode: func(value bool) (payload.Payload, error) {
		sizer := msgpack.NewSizer()
		sizer.WriteBool(value)
		buf := make([]byte, sizer.Len())
		encoder := msgpack.NewEncoder(buf)
		encoder.WriteBool(value)
		return payload.New(buf), nil
	},
}

var Time = Transform[time.Time]{
	Decode: func(raw payload.Payload) (time.Time, error) {
		decoder := msgpack.NewDecoder(raw.Data())
		val, err := decoder.ReadString()
		if err != nil {
			return time.Time{}, err
		}
		return time.Parse(time.RFC3339Nano, val)
	},
	Encode: func(value time.Time) (payload.Payload, error) {
		sizer := msgpack.NewSizer()
		val := value.Format(time.RFC3339Nano)
		sizer.WriteString(val)
		buf := make([]byte, sizer.Len())
		encoder := msgpack.NewEncoder(buf)
		encoder.WriteString(val)
		return payload.New(buf), nil
	},
}

var Void = Transform[struct{}]{
	Decode: func(raw payload.Payload) (struct{}, error) {
		return struct{}{}, nil
	},
	Encode: func(_ struct{}) (payload.Payload, error) {
		return payload.New(nil), nil
	},
}

type MsgPackCodecPtr[T any] interface {
	msgpack.Codec
	*T
}

func MsgPackDecode[T any, P MsgPackCodecPtr[T]](raw payload.Payload) (value T, err error) {
	var p P = &value
	err = CodecDecode(raw, p)
	return value, err
}

func MsgPackEncode[T any, P MsgPackCodecPtr[T]](value T) (payload.Payload, error) {
	var p P = &value
	return CodecEncode(p)
}

func CodecDecode(raw payload.Payload, value msgpack.Codec) error {
	decoder := msgpack.NewDecoder(raw.Data())
	return value.Decode(&decoder)
}

func CodecEncode(value msgpack.Codec) (payload.Payload, error) {
	sizer := msgpack.NewSizer()
	value.Encode(&sizer)
	buf := make([]byte, sizer.Len())
	encoder := msgpack.NewEncoder(buf)
	value.Encode(&encoder)
	return payload.New(buf), nil
}

func Int32Decode[T ~int32](raw payload.Payload) (T, error) {
	decoder := msgpack.NewDecoder(raw.Data())
	val, err := decoder.ReadInt32()
	return T(val), err
}

func Int32Encode[T ~int32](val T) (payload.Payload, error) {
	sizer := msgpack.NewSizer()
	sizer.WriteInt32(int32(val))
	buf := make([]byte, sizer.Len())
	encoder := msgpack.NewEncoder(buf)
	encoder.WriteInt32(int32(val))
	return payload.New(buf), nil
}

func Int64Decode[T ~int64](raw payload.Payload) (T, error) {
	decoder := msgpack.NewDecoder(raw.Data())
	val, err := decoder.ReadInt64()
	return T(val), err
}

func Int64Encode[T ~int64](val T) (payload.Payload, error) {
	sizer := msgpack.NewSizer()
	sizer.WriteInt64(int64(val))
	buf := make([]byte, sizer.Len())
	encoder := msgpack.NewEncoder(buf)
	encoder.WriteInt64(int64(val))
	return payload.New(buf), nil
}

func Uint64Decode[T ~uint64](raw payload.Payload) (T, error) {
	decoder := msgpack.NewDecoder(raw.Data())
	val, err := decoder.ReadUint64()
	return T(val), err
}

func Uint64Encode[T ~uint64](val T) (payload.Payload, error) {
	sizer := msgpack.NewSizer()
	sizer.WriteUint64(uint64(val))
	buf := make([]byte, sizer.Len())
	encoder := msgpack.NewEncoder(buf)
	encoder.WriteUint64(uint64(val))
	return payload.New(buf), nil
}

func StringDecode[T ~string](raw payload.Payload) (T, error) {
	decoder := msgpack.NewDecoder(raw.Data())
	val, err := decoder.ReadString()
	return T(val), err
}

func StringEncode[T ~string](val T) (payload.Payload, error) {
	sizer := msgpack.NewSizer()
	sizer.WriteString(string(val))
	buf := make([]byte, sizer.Len())
	encoder := msgpack.NewEncoder(buf)
	encoder.WriteString(string(val))
	return payload.New(buf), nil
}

func SliceDecode[T any, P MsgPackCodecPtr[T]](raw payload.Payload) ([]T, error) {
	decoder := msgpack.NewDecoder(raw.Data())
	size, _ := decoder.ReadArraySize()
	items := make([]T, size)
	i := 0
	for size > 0 {
		size--
		var item T
		p := (P)(&item)
		p.Decode(&decoder)
		items[i] = item
		i++
	}
	return items, nil
}

func SliceEncode[T any, P MsgPackCodecPtr[T]](items []T) (payload.Payload, error) {
	sizer := msgpack.NewSizer()
	sizer.WriteArraySize(uint32(len(items)))
	for _, v := range items {
		p := P(&v)
		p.Encode(&sizer)
	}
	buf := make([]byte, sizer.Len())
	encoder := msgpack.NewEncoder(buf)
	encoder.WriteArraySize(uint32(len(items)))
	for _, v := range items {
		p := P(&v)
		p.Encode(&encoder)
	}
	return payload.New(buf), nil
}

func InterfaceEncode[T any](tracker *invoke.LiveInstances[T]) rx.Transform[T, payload.Payload] {
	return func(instance T) (payload.Payload, error) {
		handle := tracker.Put(instance)
		sizer := msgpack.NewSizer()
		sizer.WriteUint64(handle)
		buf := make([]byte, sizer.Len())
		encoder := msgpack.NewEncoder(buf)
		encoder.WriteUint64(handle)
		return payload.New(buf), nil
	}
}

func FluxToVoid[T any](f flux.Flux[T]) mono.Void {
	return mono.Create(func(sink mono.Sink[struct{}]) {
		f.Subscribe(flux.Subscribe[T]{
			OnComplete: func() {
				sink.Success(struct{}{})
			},
			OnError: func(err error) {
				sink.Error(err)
			},
		})
	})
}

func FluxToMono[T any](f flux.Flux[T]) mono.Mono[T] {
	return mono.Create(func(sink mono.Sink[T]) {
		f.Subscribe(flux.Subscribe[T]{
			OnNext: func(val T) {
				sink.Success(val)
			},
			OnError: func(err error) {
				sink.Error(err)
			},
		})
	})
}
