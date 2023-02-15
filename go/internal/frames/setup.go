package frames

import (
	"encoding/binary"
	"time"
)

// https://rsocket.io/about/protocol/#setup-frame-0x01

type Setup struct {
	MajorVersion         uint16
	MinorVersion         uint16
	TimeBetweenKeepalive time.Duration
	MaxLifetime          time.Duration
	Token                []byte
	MimeMetadata         string
	MimeData             string
	Metadata             []byte
	Data                 []byte
	Lease                bool
}

func (f *Setup) GetStreamID() uint32 {
	return 0
}

func (f *Setup) Type() FrameType {
	return FrameTypeSetup
}

func (f *Setup) Decode(header *FrameHeader, payload []byte) error {
	flags := header.Flag()
	hasMetadata := flags.Check(FlagMetadata)
	resume := flags.Check(FlagResume)
	lease := flags.Check(FlagLease)
	var metadata []byte

	major := binary.BigEndian.Uint16(payload)
	minor := binary.BigEndian.Uint16(payload[2:])
	timeBetweenKeepalive := time.Millisecond * time.Duration(binary.BigEndian.Uint32(payload[4:]))
	maxLifetime := time.Millisecond * time.Duration(binary.BigEndian.Uint32(payload[8:]))
	var token []byte

	data := payload[12:]

	if resume {
		tokenLength := binary.BigEndian.Uint16(data)
		token = data[2 : 2+tokenLength]
		data = data[2+tokenLength:]
	}

	metadataMimeLength := data[0]
	mimeMetadata := string(data[1 : 1+metadataMimeLength])
	data = data[1+metadataMimeLength:]

	dataMimeLength := data[0]
	mimeData := string(data[1 : 1+dataMimeLength])
	data = data[1+dataMimeLength:]
	payload = data

	if hasMetadata {
		var mdLengthBuf [4]byte
		copy(mdLengthBuf[1:], data[0:3])
		mdLength := binary.BigEndian.Uint32(mdLengthBuf[:])
		data = data[3:]
		metadata = data[:mdLength]
		payload = data[mdLength:]
	}

	*f = Setup{
		MajorVersion:         major,
		MinorVersion:         minor,
		TimeBetweenKeepalive: timeBetweenKeepalive,
		MaxLifetime:          maxLifetime,
		Token:                token,
		MimeMetadata:         mimeMetadata,
		MimeData:             mimeData,
		Metadata:             metadata,
		Data:                 payload,
		Lease:                lease,
	}

	return nil
}

func (f *Setup) Encode(buf []byte) {
	tokenLen := uint16(len(f.Token))
	metadataLen := uint32(len(f.Metadata))
	payload := buf
	var flags FrameFlag
	if f.Lease {
		flags |= FlagLease
	}
	if tokenLen > 0 {
		flags |= FlagResume
	}
	if metadataLen > 0 {
		flags |= FlagMetadata
	}

	ResetFrameHeader(payload, 0, FrameTypeSetup, flags)
	payload = payload[FrameHeaderLen:]

	binary.BigEndian.PutUint16(payload, f.MajorVersion)
	binary.BigEndian.PutUint16(payload[2:], f.MinorVersion)
	binary.BigEndian.PutUint32(payload[4:], uint32(f.TimeBetweenKeepalive)/uint32(time.Millisecond))
	binary.BigEndian.PutUint32(payload[8:], uint32(f.MaxLifetime)/uint32(time.Millisecond))

	payload = payload[12:]

	if tokenLen > 0 {
		binary.BigEndian.PutUint16(payload[:], tokenLen)
		payload = payload[2:]
		copy(payload, f.Token)
		payload = payload[tokenLen:]
	}

	metadataMimeBytes := []byte(f.MimeMetadata)
	payload[0] = byte(len(metadataMimeBytes))
	copy(payload[1:], metadataMimeBytes)
	payload = payload[1+len(metadataMimeBytes):]

	dataMimeBytes := []byte(f.MimeData)
	payload[0] = byte(len(dataMimeBytes))
	copy(payload[1:], dataMimeBytes)
	payload = payload[1+len(dataMimeBytes):]

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

func (f *Setup) Size() uint32 {
	size := uint32(FrameHeaderLen) + 14

	metadataLen := uint32(len(f.Metadata))
	tokenLen := uint32(len(f.Token))
	size += metadataLen + tokenLen + uint32(len(f.Data)+len(f.Token)+len(f.MimeMetadata)+len(f.MimeData))
	if tokenLen > 0 {
		size += 2
	}
	if metadataLen > 0 {
		size += 3
	}
	return size
}
