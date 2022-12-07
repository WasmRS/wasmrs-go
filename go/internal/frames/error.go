package frames

import (
	"encoding/binary"
)

// https://rsocket.io/about/protocol#error-frame-0x0b

type Error struct {
	StreamID uint32
	Code     ErrCode
	Data     string
}

func (f *Error) GetStreamID() uint32 {
	return f.StreamID
}

func (f *Error) Type() FrameType {
	return FrameTypeError
}

func (f *Error) Decode(header *FrameHeader, payload []byte) error {
	code := binary.BigEndian.Uint32(payload)
	data := string(payload[4:])

	*f = Error{
		StreamID: header.StreamID(),
		Code:     ErrCode(code),
		Data:     data,
	}

	return nil
}

func (f *Error) Encode(buf []byte) {
	payload := buf
	ResetFrameHeader(payload, f.StreamID, FrameTypeError, 0)
	payload = payload[FrameHeaderLen:]
	binary.BigEndian.PutUint32(payload, uint32(f.Code))
	payload = payload[4:]
	copy(payload, []byte(f.Data))
}

func (f *Error) Size() uint32 {
	return uint32(FrameHeaderLen + 4 + len(f.Data))
}
