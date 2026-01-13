package internal

import (
	"bytes"
	"encoding/gob"

	"github.com/golang/snappy"
)

func Encode(value any) []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(value)
	if err != nil {
		return nil
	}
	compressed := snappy.Encode(nil, buffer.Bytes())
	return compressed
}

func Decode[V any](data []byte) V {
	decompressed, _ := snappy.Decode(nil, data)
	buffer := bytes.NewBuffer(decompressed)
	decoder := gob.NewDecoder(buffer)
	var value V
	decoder.Decode(&value)
	return value
}
