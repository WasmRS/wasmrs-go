package frames

import (
	"encoding/binary"

	"golang.org/x/exp/constraints"
)

// https://rsocket.io/about/protocol#request_response-frame-0x04
// https://rsocket.io/about/protocol#request_fnf-fire-n-forget-frame-0x05
// https://rsocket.io/about/protocol#request_stream-frame-0x06
// https://rsocket.io/about/protocol#request_channel-frame-0x07

type RequestPayload struct {
	FrameType FrameType
	StreamID  uint32
	Metadata  []byte
	Data      []byte
	Follows   bool
	Complete  bool
	InitialN  uint32
}

func (f *RequestPayload) GetStreamID() uint32 {
	return f.StreamID
}

func (f *RequestPayload) Type() FrameType {
	return f.FrameType
}

func (f *RequestPayload) Decode(header *FrameHeader, payload []byte) error {
	frameType := header.Type()
	flags := header.Flag()
	hasMetadata := flags.Check(FlagMetadata)
	follows := flags.Check(FlagFollow)
	initialN := uint32(1)
	complete := true

	if frameType == FrameTypeRequestStream || frameType == FrameTypeRequestChannel {
		initialN = binary.BigEndian.Uint32(payload)
		payload = payload[4:]
	}
	if frameType == FrameTypeRequestChannel {
		complete = flags.Check(FlagComplete)
	}

	metadata := emptyBuffer
	if hasMetadata {
		var mdLengthBuf [4]byte
		copy(mdLengthBuf[1:], payload[0:3])
		mdLength := binary.BigEndian.Uint32(mdLengthBuf[:])
		payload = payload[3:]
		metadata = payload[:mdLength]
		payload = payload[mdLength:]
	}

	*f = RequestPayload{
		FrameType: frameType,
		StreamID:  header.StreamID(),
		Metadata:  metadata,
		Data:      payload,
		Follows:   follows,
		Complete:  complete,
		InitialN:  initialN,
	}

	return nil
}

func (f *RequestPayload) Encode(buf []byte) {
	payload := buf
	metadataLen := uint32(len(f.Metadata))
	var flags FrameFlag
	if f.FrameType == FrameTypeRequestChannel && f.Complete {
		flags |= FlagComplete
	}

	if metadataLen > 0 {
		flags |= FlagMetadata
	}

	ResetFrameHeader(payload, f.StreamID, f.FrameType, flags)
	payload = payload[FrameHeaderLen:]

	if f.FrameType == FrameTypeRequestStream || f.FrameType == FrameTypeRequestChannel {
		binary.BigEndian.PutUint32(payload, f.InitialN)
		payload = payload[4:]
	}

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

func (f *RequestPayload) Size() uint32 {
	size := uint32(FrameHeaderLen)
	if f.FrameType == FrameTypeRequestStream || f.FrameType == FrameTypeRequestChannel {
		size += 4
	}

	size += uint32(len(f.Metadata) + len(f.Data))
	metadataLen := uint32(len(f.Metadata))
	if metadataLen > 0 {
		size += 3
	}
	return size
}

func (f *RequestPayload) Fragment(maxFrameSize uint32) []Frame {
	size := f.Size()
	if size <= maxFrameSize {
		return nil
	}

	var frames []Frame

	metadata := f.Metadata
	data := f.Data

	nonDataSize := uint32(FrameHeaderLen)
	if f.FrameType == FrameTypeRequestStream || f.FrameType == FrameTypeRequestChannel {
		nonDataSize += 4
	}

	maxMDSize := maxFrameSize - nonDataSize - 3
	lenMd := uint32(len(metadata))
	if lenMd >= maxMDSize {
		metadataFrag := metadata[:maxMDSize]
		metadata = metadata[maxMDSize:]
		frames = append(frames, &RequestPayload{
			FrameType: f.FrameType,
			StreamID:  f.StreamID,
			Metadata:  metadataFrag,
			Data:      nil,
			Follows:   true,
			Complete:  false,
			InitialN:  f.InitialN,
		})
	} else {
		diff := size - maxFrameSize
		dataOff := uint32(len(data)) - diff
		dataFrag := data[:dataOff]
		data = data[dataOff:]
		frames = append(frames, &RequestPayload{
			FrameType: f.FrameType,
			StreamID:  f.StreamID,
			Metadata:  metadata,
			Data:      dataFrag,
			Follows:   true,
			Complete:  false,
			InitialN:  f.InitialN,
		})
	}

	frames = splitPayloads(frames, f.StreamID, true, f.Complete, maxFrameSize, data, metadata)

	return frames
}

func min[T constraints.Ordered](left, right T) T {
	if left < right {
		return left
	}
	return right
}
