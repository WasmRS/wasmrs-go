package frames

import (
	"encoding/binary"
)

// https://rsocket.io/about/protocol#payload-frame-0x0a

type Payload struct {
	StreamID uint32
	Metadata []byte
	Data     []byte
	Follows  bool
	Complete bool
	Next     bool
}

func (f *Payload) Type() FrameType {
	return FrameTypePayload
}

func (f *Payload) Decode(header *FrameHeader, payload []byte) error {
	flags := header.Flag()
	hasMetadata := flags.Check(FlagMetadata)
	follows := flags.Check(FlagFollow)
	complete := flags.Check(FlagComplete)
	next := flags.Check(FlagNext)
	metadata := emptyBuffer

	if hasMetadata {
		var mdLengthBuf [4]byte
		copy(mdLengthBuf[1:], payload[0:3])
		mdLength := binary.BigEndian.Uint32(mdLengthBuf[:])
		payload = payload[3:]
		metadata = payload[:mdLength]
		payload = payload[mdLength:]
	}

	*f = Payload{
		StreamID: header.StreamID(),
		Metadata: metadata,
		Data:     payload,
		Follows:  follows,
		Complete: complete,
		Next:     next,
	}

	return nil
}

func (f *Payload) Encode(buf []byte) {
	size := FrameHeaderLen
	size += len(f.Metadata) + len(f.Data)
	metadataLen := uint32(len(f.Metadata))
	if metadataLen > 0 {
		size += 3
	}

	payload := buf
	var flags FrameFlag
	if f.Follows {
		flags |= FlagFollow
	}
	if f.Complete {
		flags |= FlagComplete
	}
	if f.Next {
		flags |= FlagNext
	}
	if metadataLen > 0 {
		flags |= FlagMetadata
	}

	ResetFrameHeader(payload, f.StreamID, FrameTypePayload, flags)
	payload = payload[FrameHeaderLen:]

	if metadataLen > 0 {
		var mdLengthBuf [4]byte
		binary.BigEndian.PutUint32(mdLengthBuf[:], metadataLen)
		copy(payload[0:3], mdLengthBuf[1:])
		payload = payload[3:]
		copy(payload, f.Metadata)
		payload = payload[metadataLen:]
	}

	copy(payload, f.Data)
}

func (f *Payload) Size() uint32 {
	size := uint32(FrameHeaderLen)

	size += uint32(len(f.Metadata) + len(f.Data))
	metadataLen := uint32(len(f.Metadata))
	if metadataLen > 0 {
		size += 3
	}
	return size
}

func (f *Payload) Fragment(maxFrameSize uint32) []Frame {
	size := f.Size()
	if size <= maxFrameSize {
		return nil
	}

	return splitPayloads(nil, f.StreamID, f.Next, f.Complete, maxFrameSize, f.Data, f.Metadata)
}
