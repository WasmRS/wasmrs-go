package frames

type Frame interface {
	Type() FrameType
	Decode(header *FrameHeader, payload []byte) error
	Size() uint32
	Encode(buffer []byte)
}

type Fragmentable interface {
	Frame
	Fragment(maxFrameSize uint32) []Frame
}
