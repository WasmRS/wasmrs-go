package payload

type Payload interface {
	Metadata() []byte
	Data() []byte
}

type SimplePayload struct {
	metadata []byte
	data     []byte
}

func New(data []byte, metadata ...[]byte) *SimplePayload {
	var md []byte
	if len(metadata) > 0 {
		md = metadata[0]
	}
	return &SimplePayload{
		metadata: md,
		data:     data,
	}
}

func (p *SimplePayload) Metadata() []byte {
	return p.metadata
}

func (p *SimplePayload) Data() []byte {
	return p.data
}
