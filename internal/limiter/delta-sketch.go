package limiter

import (
	"bytes"
	"encoding/binary"
)

type EncodedDeltaSketch struct {
	UsedTokens   []byte
	UsableTokens []byte
	LastRefilled []byte
}

func (ds *EncodedDeltaSketch) encode() []byte {
	buf := new(bytes.Buffer)
	writeField := func(field []byte) {
		binary.Write(buf, binary.LittleEndian, uint32(len(field)))
		buf.Write(field)
	}
	writeField(ds.UsedTokens)
	writeField(ds.UsableTokens)
	writeField(ds.LastRefilled)
	return buf.Bytes()
}

func decode(payload []byte) *EncodedDeltaSketch {
	buf := bytes.NewReader(payload)
	readField := func() []byte {
		var length uint32
		binary.Read(buf, binary.LittleEndian, &length)
		field := make([]byte, length)
		buf.Read(field)
		return field
	}
	return &EncodedDeltaSketch{
		UsedTokens:   readField(),
		UsableTokens: readField(),
		LastRefilled: readField(),
	}
}
