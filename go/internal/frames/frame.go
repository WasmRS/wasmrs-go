package frames

type Frame interface {
	GetStreamID() uint32
	Type() FrameType
	Decode(header *FrameHeader, payload []byte) error
	Size() uint32
	Encode(buffer []byte)
}

type Fragmentable interface {
	Frame
	Fragment(maxFrameSize uint32) []Frame
}
