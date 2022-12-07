package frames

import (
	"encoding/binary"
)

var emptyBuffer = []byte{}

func readString(payload []byte) (string, []byte) {
	strLen := binary.BigEndian.Uint16(payload[:2])
	payload = payload[2:]
	strBytes := payload[:strLen]
	str := string(strBytes)
	return str, payload[strLen:]
}

func splitPayloads(frames []Frame, streamID uint32, next, complete bool, maxFrameSize uint32, data, metadata []byte) []Frame {
	for {
		lenMd := uint32(len(metadata))
		lenData := uint32(len(data))
		if lenMd+lenData == 0 {
			break
		}

		nonDataSize := uint32(FrameHeaderLen)
		if lenMd > 0 {
			nonDataSize += 3
		}

		space := maxFrameSize - nonDataSize
		follows := nonDataSize+lenMd+lenData > maxFrameSize
		metadataFrag := metadata
		dataFrag := data

		if lenMd > space {
			dataFrag = dataFrag[:0] // len = 0
			metadataFrag = metadataFrag[:space]
			metadata = metadata[space:]
		} else {
			metadata = metadataFrag[lenMd:lenMd] // len = 0
			if lenData > space-lenMd {
				dataSpace := min(space-lenMd, lenData)
				dataFrag = dataFrag[:dataSpace]
				data = data[dataSpace:]
			} else {
				data = data[lenData:lenData] // len = 0
			}
		}

		frames = append(frames, &Payload{
			StreamID: streamID,
			Metadata: metadataFrag,
			Data:     dataFrag,
			Follows:  follows,
			Complete: !follows && complete,
			Next:     !follows && next,
		})
	}

	return frames
}
