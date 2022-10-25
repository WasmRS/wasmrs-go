package customer

import (
	"github.com/nanobus/iota/go/msgpack"
)

type Customer struct {
	FirstName string `json:"firstName" msgpack:"firstName"`
	LastName  string `json:"lastName" msgpack:"lastName"`
}

func (s *Customer) Decode(decoder *msgpack.Decoder) error {
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
		case "firstName":
			s.FirstName, err = decoder.ReadString()
		case "lastName":
			s.LastName, err = decoder.ReadString()
		default:
			err = decoder.Skip()
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Customer) Encode(encoder msgpack.Writer) error {
	if s == nil {
		encoder.WriteNil()
		return nil
	}
	encoder.WriteMapSize(2)
	encoder.WriteString("firstName")
	encoder.WriteString(s.FirstName)
	encoder.WriteString("lastName")
	encoder.WriteString(s.LastName)
	return nil
}
