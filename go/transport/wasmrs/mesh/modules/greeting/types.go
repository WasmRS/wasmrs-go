package greeting

import (
	"github.com/nanobus/iota/go/msgpack"
)

type GreetingRequest struct {
	Name string `json:"name" msgpack:"name"`
}

func (r *GreetingRequest) Decode(decoder msgpack.Reader) error {
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
		case "name":
			r.Name, err = decoder.ReadString()
		default:
			err = decoder.Skip()
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *GreetingRequest) Encode(encoder msgpack.Writer) error {
	if r == nil {
		encoder.WriteNil()
		return nil
	}
	encoder.WriteMapSize(1)
	encoder.WriteString("name")
	encoder.WriteString(r.Name)
	return nil
}
