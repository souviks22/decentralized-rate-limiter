package utils

import (
	"bytes"
	"encoding/gob"

	"github.com/golang/snappy"
)

func Encode(value any) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(value)
	if err != nil {
		return nil
	}
	compressed := snappy.Encode(nil, buf.Bytes())
	return compressed
}

func Decode[V any](data []byte) V {
	decompressed, _ := snappy.Decode(nil, data)
	var value V
	buf := bytes.NewBuffer(decompressed)
	dec := gob.NewDecoder(buf)
	dec.Decode(&value)
	return value
}
