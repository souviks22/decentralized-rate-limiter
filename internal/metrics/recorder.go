package metrics

import (
	"log"
	"os"
	"sort"
	"sync"
)

var enabled = os.Getenv("METRICS_ENABLED") == "true"

type Recorder struct {
	mu      sync.Mutex
	samples []float64
	sum     float64
	max     float64
	name    string
	unit    string
	scale   func(any) float64
}

func NewRecorder(name, unit string, scale func(any) float64) *Recorder {
	return &Recorder{name: name, unit: unit, scale: scale}
}

func (r *Recorder) Record(v any) {
	if !enabled {
		return
	}
	rv := r.scale(v)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.samples = append(r.samples, rv)
	r.sum += rv
	if rv > r.max {
		r.max = rv
	}
}

func (r *Recorder) snapshotAndReset() (mean, p50, p95, p99, max float64, count int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := len(r.samples)
	if n == 0 {
		return 0, 0, 0, 0, 0, 0
	}
	sorted := make([]float64, n)
	copy(sorted, r.samples)
	sort.Float64s(sorted)
	count = n
	mean = r.sum / float64(n)
	p50 = sorted[int(0.50*float64(n))]
	p95 = sorted[int(0.95*float64(n))]
	p99 = sorted[int(0.99*float64(n))]
	max = r.max
	r.samples = nil
	r.sum = 0
	r.max = 0
	return
}

func (r *Recorder) LogSnapshot() {
	if !enabled {
		return
	}
	mean, p50, p95, p99, max, count := r.snapshotAndReset()
	if count == 0 {
		return
	}
	log.Printf("[%s] count=%d mean=%.2f%s p50=%.2f%s p95=%.2f%s p99=%.2f%s max=%.2f%s",
		r.name, count, mean, r.unit, p50, r.unit, p95, r.unit, p99, r.unit, max, r.unit)
}
