// Package metrics — Prometheus adapter.
//
// PrometheusRecorder implements [MetricsRecorder] by registering and updating
// Prometheus collector types (Counter, Gauge, Histogram).
package metrics

import "time"

// PrometheusRecorder implements [MetricsRecorder] backed by Prometheus client_golang.
type PrometheusRecorder struct {
	// TODO: hold prometheus.Registerer + cached metric vectors
}

// NewPrometheusRecorder creates a [PrometheusRecorder] registered with the
// default Prometheus registry.
func NewPrometheusRecorder() *PrometheusRecorder {
	return &PrometheusRecorder{}
}

// Counter increments a counter metric by the given value.
func (p *PrometheusRecorder) Counter(name string, value float64, tags map[string]string) {
	// TODO: get-or-create CounterVec, .With(tags).Add(value)
}

// Gauge sets a gauge metric to the given value.
func (p *PrometheusRecorder) Gauge(name string, value float64, tags map[string]string) {
	// TODO: get-or-create GaugeVec, .With(tags).Set(value)
}

// Histogram records a single observation in a distribution.
func (p *PrometheusRecorder) Histogram(name string, value float64, tags map[string]string) {
	// TODO: get-or-create HistogramVec, .With(tags).Observe(value)
}

// Timer starts a timer and returns a stop function that records the elapsed duration.
func (p *PrometheusRecorder) Timer(name string) func() {
	start := time.Now()
	return func() {
		elapsed := time.Since(start).Seconds()
		p.Histogram(name, elapsed, nil)
	}
}

var _ Recorder = (*PrometheusRecorder)(nil)
