package invoke

import (
	"encoding/binary"
	"strings"
)

type PortIO struct {
	Port     string
	Data     []byte
	Complete bool
	Next     bool
}

func (f *PortIO) Decode(data []byte) error {
	flags := Flags(data[0])
	complete := flags.Check(FlagComplete)
	next := flags.Check(FlagNext)
	data = data[1:]
	portLength := binary.BigEndian.Uint16(data)
	data = data[2:]
	port := string(data[:portLength])
	data = data[portLength:]

	*f = PortIO{
		Port:     port,
		Data:     data,
		Complete: complete,
		Next:     next,
	}

	return nil
}

func (f *PortIO) Encode() []byte {
	size := 3 + len(f.Port) + len(f.Data)
	payload := make([]byte, size)
	buf := payload
	var flags Flags
	if f.Complete {
		flags |= FlagComplete
	}
	if f.Next {
		flags |= FlagNext
	}

	payload[0] = byte(flags)
	payload = payload[1:]
	binary.BigEndian.PutUint16(payload, uint16(len(f.Port)))
	payload = payload[2:]
	copy(payload, []byte(f.Port))
	payload = payload[len(f.Port):]
	copy(payload, f.Data)

	return buf
}

type Flags uint8

func (f Flags) String() string {
	var _foo [2]string
	foo := _foo[:0]
	if f.Check(FlagNext) {
		foo = append(foo, "N")
	}
	if f.Check(FlagComplete) {
		foo = append(foo, "CL")
	}
	return strings.Join(foo, "|")
}

// All frame flags
const (
	FlagNext Flags = 1
	FlagComplete
)

// Check returns true if mask exists.
func (f Flags) Check(flag Flags) bool {
	return flag&f == flag
}
