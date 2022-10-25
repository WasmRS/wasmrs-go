package frames

import (
	"encoding/binary"
)

// https://rsocket.io/about/protocol#request_n-frame-0x08

type RequestN struct {
	StreamID uint32
	N        uint32
}

func (f *RequestN) Type() FrameType {
	return FrameTypeRequestN
}

func (f *RequestN) Decode(header *FrameHeader, payload []byte) error {
	n := binary.BigEndian.Uint32(payload)

	*f = RequestN{
		StreamID: header.StreamID(),
		N:        n,
	}

	return nil
}

func (f *RequestN) Encode(buf []byte) {
	payload := buf
	ResetFrameHeader(payload, f.StreamID, FrameTypeRequestN, 0)
	payload = payload[FrameHeaderLen:]
	binary.BigEndian.PutUint32(payload, f.N)
}

func (f *RequestN) Size() uint32 {
	return FrameHeaderLen + 4
}
