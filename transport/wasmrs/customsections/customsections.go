package customsections

import (
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/tetratelabs/wabin/binary"
	"github.com/tetratelabs/wabin/leb128"
	"github.com/tetratelabs/wabin/wasm"
)

// version is format version and doesn't change between known specification versions
// See https://www.w3.org/TR/2019/REC-wasm-core-1-20191205/#binary-version
var version = []byte{0x01, 0x00, 0x00, 0x00}

// Read implements wasm.DecodeModule for the WebAssembly Binary Format
// See https://www.w3.org/TR/2019/REC-wasm-core-1-20191205/#binary-format%E2%91%A0
func Read(data []byte, names ...string) (map[string]*wasm.CustomSection, error) {
	r := bytes.NewReader(data)

	// Magic number.
	buf := make([]byte, 4)
	if _, err := io.ReadFull(r, buf); err != nil || !bytes.Equal(buf, binary.Magic) {
		return nil, binary.ErrInvalidMagicNumber
	}

	// Version.
	if _, err := io.ReadFull(r, buf); err != nil || !bytes.Equal(buf, version) {
		return nil, binary.ErrInvalidVersion
	}

	wantedNames := make(map[string]struct{})
	for _, name := range names {
		wantedNames[name] = struct{}{}
	}
	sections := make(map[string]*wasm.CustomSection, len(wantedNames))

	for {
		// TODO: except custom sections, all others are required to be in order, but we aren't checking yet.
		// See https://www.w3.org/TR/2019/REC-wasm-core-1-20191205/#modules%E2%91%A0%E2%93%AA
		sectionID, err := r.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("read section id: %w", err)
		}

		sectionSize, _, err := leb128.DecodeUint32(r)
		if err != nil {
			return nil, fmt.Errorf("get size of section %s: %v", wasm.SectionIDName(sectionID), err)
		}

		sectionContentStart := r.Len()
		switch sectionID {
		case wasm.SectionIDCustom:
			// First, validate the section and determine if the section for this name has already been set
			name, nameSize, decodeErr := decodeUTF8(r, "custom section name")
			if decodeErr != nil {
				err = decodeErr
				break
			} else if sectionSize < nameSize {
				err = fmt.Errorf("malformed custom section %s", name)
				break
			}

			limit := sectionSize - nameSize
			if _, ok := wantedNames[name]; ok {
				custom, err := decodeCustomSection(r, name, uint64(limit))
				if err != nil {
					return nil, fmt.Errorf("failed to read custom section name[%s]: %w", name, err)
				}
				sections[name] = custom
			} else {
				if _, err = io.CopyN(io.Discard, r, int64(limit)); err != nil {
					return nil, fmt.Errorf("failed to skip section: %w", err)
				}
			}

		default:
			// Note: Not Seek because it doesn't err when given an offset past EOF. Rather, it leads to undefined state.
			if _, err = io.CopyN(io.Discard, r, int64(sectionSize)); err != nil {
				return nil, fmt.Errorf("failed to skip section: %w", err)
			}
		}

		readBytes := sectionContentStart - r.Len()
		if err == nil && int(sectionSize) != readBytes {
			err = fmt.Errorf("invalid section length: expected to be %d but got %d", sectionSize, readBytes)
		}

		if err != nil {
			return nil, fmt.Errorf("section %s: %v", wasm.SectionIDName(sectionID), err)
		}
	}

	return sections, nil
}

// decodeUTF8 decodes a size prefixed string from the reader, returning it and the count of bytes read.
// contextFormat and contextArgs apply an error format when present
func decodeUTF8(r *bytes.Reader, contextFormat string, contextArgs ...interface{}) (string, uint32, error) {
	size, sizeOfSize, err := leb128.DecodeUint32(r)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read %s size: %w", fmt.Sprintf(contextFormat, contextArgs...), err)
	}

	buf := make([]byte, size)
	if _, err = io.ReadFull(r, buf); err != nil {
		return "", 0, fmt.Errorf("failed to read %s: %w", fmt.Sprintf(contextFormat, contextArgs...), err)
	}

	if !utf8.Valid(buf) {
		return "", 0, fmt.Errorf("%s is not valid UTF-8", fmt.Sprintf(contextFormat, contextArgs...))
	}

	return string(buf), size + uint32(sizeOfSize), nil
}

// decodeCustomSection deserializes the data **not** associated with the "name" key in SectionIDCustom.
//
// See https://www.w3.org/TR/2019/REC-wasm-core-1-20191205/#custom-section%E2%91%A0
func decodeCustomSection(r *bytes.Reader, name string, limit uint64) (result *wasm.CustomSection, err error) {
	buf := make([]byte, limit)
	_, err = r.Read(buf)

	result = &wasm.CustomSection{
		Name: name,
		Data: buf,
	}

	return
}
