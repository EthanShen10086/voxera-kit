// Package metrics — Prometheus adapter.
//
// PrometheusRecorder implements [Recorder] by registering and updating
// Prometheus collector types (Counter, Gauge, Histogram).
package metrics

import (
	"sort"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusRecorder implements [Recorder] backed by the Prometheus client
// library. Metric vectors are lazily created and cached for reuse.
type PrometheusRecorder struct {
	registerer prometheus.Registerer

	mu         sync.Mutex
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	histograms map[string]*prometheus.HistogramVec
}

// NewPrometheusRecorder creates a [PrometheusRecorder] registered with the
// default Prometheus registry.
func NewPrometheusRecorder() *PrometheusRecorder {
	return &PrometheusRecorder{
		registerer: prometheus.DefaultRegisterer,
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
	}
}

func labelNames(tags map[string]string) []string {
	names := make([]string, 0, len(tags))
	for k := range tags {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func (p *PrometheusRecorder) getOrCreateCounter(name string, tags map[string]string) *prometheus.CounterVec {
	p.mu.Lock()
	defer p.mu.Unlock()

	if cv, ok := p.counters[name]; ok {
		return cv
	}
	cv := prometheus.NewCounterVec(prometheus.CounterOpts{Name: name}, labelNames(tags))
	p.registerer.MustRegister(cv)
	p.counters[name] = cv
	return cv
}

func (p *PrometheusRecorder) getOrCreateGauge(name string, tags map[string]string) *prometheus.GaugeVec {
	p.mu.Lock()
	defer p.mu.Unlock()

	if gv, ok := p.gauges[name]; ok {
		return gv
	}
	gv := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: name}, labelNames(tags))
	p.registerer.MustRegister(gv)
	p.gauges[name] = gv
	return gv
}

func (p *PrometheusRecorder) getOrCreateHistogram(name string, tags map[string]string) *prometheus.HistogramVec {
	p.mu.Lock()
	defer p.mu.Unlock()

	if hv, ok := p.histograms[name]; ok {
		return hv
	}
	hv := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name,
		Buckets: prometheus.DefBuckets,
	}, labelNames(tags))
	p.registerer.MustRegister(hv)
	p.histograms[name] = hv
	return hv
}

// Counter increments a counter metric by the given value.
func (p *PrometheusRecorder) Counter(name string, value float64, tags map[string]string) {
	cv := p.getOrCreateCounter(name, tags)
	cv.With(prometheus.Labels(tags)).Add(value)
}

// Gauge sets a gauge metric to the given value.
func (p *PrometheusRecorder) Gauge(name string, value float64, tags map[string]string) {
	gv := p.getOrCreateGauge(name, tags)
	gv.With(prometheus.Labels(tags)).Set(value)
}

// Histogram records a single observation in a distribution.
func (p *PrometheusRecorder) Histogram(name string, value float64, tags map[string]string) {
	hv := p.getOrCreateHistogram(name, tags)
	hv.With(prometheus.Labels(tags)).Observe(value)
}

// Timer starts a timer and returns a stop function that records the elapsed
// duration as a histogram observation in seconds.
func (p *PrometheusRecorder) Timer(name string) func() {
	start := time.Now()
	return func() {
		elapsed := time.Since(start).Seconds()
		p.Histogram(name, elapsed, nil)
	}
}

var _ Recorder = (*PrometheusRecorder)(nil)
