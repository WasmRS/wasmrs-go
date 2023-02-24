package operations

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strconv"
)

type Operation struct {
	Index     uint32
	Type      RequestType
	Direction Direction
	Namespace string
	Operation string
}

type Table []Operation

type RequestType uint8

const (
	RequestResponse RequestType = 1
	FireAndForget   RequestType = 2
	RequestStream   RequestType = 3
	RequestChannel  RequestType = 4
)

func (r RequestType) String() string {
	switch r {
	case RequestResponse:
		return "RequestResponse"
	case FireAndForget:
		return "FireAndForget"
	case RequestStream:
		return "RequestStream"
	case RequestChannel:
		return "RequestChannel"
	default:
		return "unknown"
	}
}

type Direction uint8

const (
	Export Direction = 1
	Import Direction = 2
)

func (d Direction) String() string {
	switch d {
	case Export:
		return "export"
	case Import:
		return "import"
	default:
		return "unknown"
	}
}

// Magic is the 4 byte preamble (literally "\0wrs") of the binary format
var Magic = []byte{0x00, 0x77, 0x72, 0x73}
var ErrInvalidMagicNumber = errors.New("invalid magic number")

func FromBytes(buf []byte) (Table, error) {
	if !bytes.Equal(buf[0:4], Magic) {
		return nil, ErrInvalidMagicNumber
	}
	buf = buf[4:]

	// Read version.
	version := binary.BigEndian.Uint16(buf)
	if version != 1 {
		return nil, errors.New("Unknown operation table version %d" + strconv.Itoa(int(version)))
	}
	buf = buf[2:]
	numOps := binary.BigEndian.Uint32(buf)
	operations := make(Table, numOps)
	buf = buf[4:]
	for i := 0; i < int(numOps); i++ {
		opType := RequestType(buf[0])
		dir := Direction(buf[1])
		id := binary.BigEndian.Uint32(buf[2:6])
		nsLen := binary.BigEndian.Uint16(buf[6:8])
		buf = buf[8:]
		ns := string(buf[:nsLen])
		buf = buf[nsLen:]
		opLen := binary.BigEndian.Uint16(buf[0:2])
		buf = buf[2:]
		op := string(buf[:opLen])
		buf = buf[opLen:]
		operations[i] = Operation{
			Type:      opType,
			Direction: dir,
			Index:     id,
			Namespace: ns,
			Operation: op,
		}
		reservedLen := binary.BigEndian.Uint16(buf)
		buf = buf[2+reservedLen:] // Reserved
	}

	return operations, nil
}

func (t Table) ToBytes() []byte {
	size := uint32(10)
	for _, op := range t {
		size += uint32(12 + len(op.Namespace) + len(op.Operation))
	}

	payload := make([]byte, size)
	buf := payload
	copy(buf, Magic)
	buf = buf[4:]
	binary.BigEndian.PutUint16(buf, 1)
	buf = buf[2:]
	binary.BigEndian.PutUint32(buf, uint32(len(t)))
	buf = buf[4:]

	for _, oper := range t {
		buf[0] = byte(oper.Type)
		buf[1] = byte(oper.Direction)

		// Operation ID (index)
		binary.BigEndian.PutUint32(buf[2:6], oper.Index)
		buf = buf[6:]

		// Namespace length and string value
		nsLen := uint16(len(oper.Namespace))
		binary.BigEndian.PutUint16(buf[0:2], nsLen)
		buf = buf[2:]
		copy(buf, []byte(oper.Namespace))
		buf = buf[nsLen:]

		// Operation length and string value
		operLen := uint16(len(oper.Operation))
		binary.BigEndian.PutUint16(buf[0:2], operLen)
		buf = buf[2:]
		copy(buf, []byte(oper.Operation))
		buf = buf[operLen:]

		// Reserved
		binary.BigEndian.PutUint16(buf, 0)
		buf = buf[2:]
	}

	return payload
}
