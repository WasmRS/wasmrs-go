package concat

import (
	"github.com/nanobus/iota/go/msgpack"
)

type Strings struct {
	Left  string `json:"left" msgpack:"left"`
	Right string `json:"right" msgpack:"right"`
}

func (s *Strings) Decode(decoder msgpack.Reader) error {
	numFields, err := decoder.ReadMapSize()
	if err != nil {
		return err
	}

	for numFields > 0 {
		numFields--
		field, err := decoder.ReadString()
		if err != nil {
			return err
		}
		switch field {
		case "left":
			s.Left, err = decoder.ReadString()
		case "right":
			s.Right, err = decoder.ReadString()
		default:
			err = decoder.Skip()
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Strings) Encode(encoder msgpack.Writer) error {
	if s == nil {
		encoder.WriteNil()
		return nil
	}
	encoder.WriteMapSize(2)
	encoder.WriteString("left")
	encoder.WriteString(s.Left)
	encoder.WriteString("right")
	encoder.WriteString(s.Right)
	return nil
}
