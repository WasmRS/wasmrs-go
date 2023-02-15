package msgpack

import (
	"golang.org/x/exp/constraints"
)

func WriteSlice[T any](encoder Writer, values []T, valF func(Writer, T)) error {
	encoder.WriteArraySize(uint32(len(values)))
	for _, item := range values {
		valF(encoder, item)
	}

	return encoder.Err()
}

func ReadSlice[T any](reader Reader, valF func(reader Reader) (T, error)) ([]T, error) {
	listSize, err := reader.ReadArraySize()
	if err != nil {
		return nil, err
	}
	request := make([]T, 0, listSize)
	for listSize > 0 {
		listSize--
		item, err := valF(reader)
		if err != nil {
			return nil, err
		}
		request = append(request, item)
	}

	return request, nil
}

func WriteMap[K constraints.Ordered, V any](writer Writer,
	m map[K]V, keyF func(Writer, K),
	valF func(Writer, V)) error {
	writer.WriteMapSize(uint32(len(m)))
	for key, val := range m {
		keyF(writer, key)
		valF(writer, val)
	}

	return writer.Err()
}

func ReadMap[K constraints.Ordered, V any](reader Reader,
	keyF func(reader Reader) (K, error),
	valF func(reader Reader) (V, error)) (map[K]V, error) {
	mapSize, err := reader.ReadMapSize()
	if err != nil {
		return nil, err
	}
	m := make(map[K]V, mapSize)
	for mapSize > 0 {
		mapSize--
		key, err := keyF(reader)
		if err != nil {
			return nil, err
		}
		value, err := valF(reader)
		if err != nil {
			return nil, err
		}
		m[key] = value
	}

	return m, nil
}
