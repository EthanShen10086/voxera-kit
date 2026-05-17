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

func (p *PrometheusRecorder) Counter(name string, value float64, tags map[string]string) {
	// TODO: get-or-create CounterVec, .With(tags).Add(value)
}

func (p *PrometheusRecorder) Gauge(name string, value float64, tags map[string]string) {
	// TODO: get-or-create GaugeVec, .With(tags).Set(value)
}

func (p *PrometheusRecorder) Histogram(name string, value float64, tags map[string]string) {
	// TODO: get-or-create HistogramVec, .With(tags).Observe(value)
}

func (p *PrometheusRecorder) Timer(name string) func() {
	start := time.Now()
	return func() {
		elapsed := time.Since(start).Seconds()
		p.Histogram(name, elapsed, nil)
	}
}

var _ MetricsRecorder = (*PrometheusRecorder)(nil)
