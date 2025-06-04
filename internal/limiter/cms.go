package limiter

import (
	"bytes"
	"encoding/binary"
	"hash/fnv"
	"math"

	"github.com/golang/snappy"
)

const Epsilon = 0.001
const Delta = 0.001

var width = uint32(math.Ceil(math.E / Epsilon))
var depth = uint32(math.Ceil(math.Log(1 / Delta)))
var seeds = []uint32{73856093, 19349663, 83492791, 49979687, 1640531527, 2654435761, 15485863}

type CountMinSketch struct {
	sketch [][]float64
	delta  map[uint32]float64
}

func newCountMinSketch(value float64) *CountMinSketch {
	sketch := make([][]float64, depth)
	for i := range sketch {
		sketch[i] = make([]float64, width)
		for j := range sketch[i] {
			sketch[i][j] = value
		}
	}
	return &CountMinSketch{
		sketch: sketch,
		delta:  make(map[uint32]float64),
	}
}

func hash(key string, seed uint32) uint32 {
	h := fnv.New32a()
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, seed)
	h.Write(buf)
	h.Write([]byte(key))
	return h.Sum32()
}

func (cms *CountMinSketch) add(key string, inc float64) {
	for row := range depth {
		col := hash(key, seeds[row]) % width
		cms.sketch[row][col] += inc
		cms.delta[row*width+col] = cms.sketch[row][col]
	}
}

func (cms *CountMinSketch) estimate(key string) float64 {
	min := math.MaxFloat64
	for i := range depth {
		idx := hash(key, seeds[i]) % width
		min = math.Min(cms.sketch[i][idx], min)
	}
	return min
}

func (cms *CountMinSketch) merge(payload []byte) {
	incoming := decodeDelta(payload)
	for cell := range incoming {
		row := cell / width
		col := cell % width
		cms.sketch[row][col] = math.Max(incoming[cell], cms.sketch[row][col])
	}
}

func (cms *CountMinSketch) encodeDelta() []byte {
	buf := new(bytes.Buffer)
	for k, v := range cms.delta {
		binary.Write(buf, binary.LittleEndian, k)
		binary.Write(buf, binary.LittleEndian, v)
	}
	cms.delta = make(map[uint32]float64)
	return snappy.Encode(nil, buf.Bytes())
}

func decodeDelta(payload []byte) map[uint32]float64 {
	delta := make(map[uint32]float64)
	decoded, _ := snappy.Decode(nil, payload)
	buf := bytes.NewReader(decoded)
	for {
		var key uint32
		var val float64
		if binary.Read(buf, binary.LittleEndian, &key) != nil {
			break
		}
		if binary.Read(buf, binary.LittleEndian, &val) != nil {
			break
		}
		delta[key] = val
	}
	return delta
}
