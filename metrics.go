package drl

import (
	"log"
	"os"
	"sort"
	"sync"
)

var enabled = os.Getenv("METRICS_ENABLED") == "true"

type Recorder struct {
	Samples []float64         `json:"samples"`
	Sum     float64           `json:"sum"`
	Max     float64           `json:"max"`
	Name    string            `json:"name"`
	Unit    string            `json:"unit"`
	Scale   func(any) float64 `json:"scale"`
	mutex   sync.Mutex
}

func NewRecorder(name, unit string, scale func(any) float64) *Recorder {
	return &Recorder{Name: name, Unit: unit, Scale: scale}
}

func (recorder *Recorder) Record(v any) {
	if !enabled {
		return
	}
	currentValue := recorder.Scale(v)
	recorder.mutex.Lock()
	defer recorder.mutex.Unlock()
	recorder.Samples = append(recorder.Samples, currentValue)
	recorder.Sum += currentValue
	if currentValue > recorder.Max {
		recorder.Max = currentValue
	}
}

func (recorder *Recorder) snapshotAndReset() (mean, p50, p95, p99, max float64, count int) {
	recorder.mutex.Lock()
	defer recorder.mutex.Unlock()
	count = len(recorder.Samples)
	if count == 0 {
		return 0, 0, 0, 0, 0, 0
	}
	sorted := make([]float64, count)
	copy(sorted, recorder.Samples)
	sort.Float64s(sorted)
	mean = recorder.Sum / float64(count)
	p50 = sorted[int(0.50*float64(count))]
	p95 = sorted[int(0.95*float64(count))]
	p99 = sorted[int(0.99*float64(count))]
	max = recorder.Max
	recorder.Samples, recorder.Sum, recorder.Max = nil, 0, 0
	return
}

func (recorder *Recorder) LogSnapshot() {
	if !enabled {
		return
	}
	mean, p50, p95, p99, max, count := recorder.snapshotAndReset()
	if count == 0 {
		return
	}
	log.Printf("[%s] count=%d mean=%.2f%s p50=%.2f%s p95=%.2f%s p99=%.2f%s max=%.2f%s",
		recorder.Name, count, mean, recorder.Unit, p50, recorder.Unit, p95, recorder.Unit, p99, recorder.Unit, max, recorder.Unit)
}
