// Package metrics defines the port for recording application metrics
// (counters, gauges, histograms, and timers).
package metrics

// Recorder is the interface every metrics backend must implement.
type Recorder interface {
	// Counter increments a counter metric by the given value.
	Counter(name string, value float64, tags map[string]string)

	// Gauge sets a gauge metric to the given value.
	Gauge(name string, value float64, tags map[string]string)

	// Histogram records a single observation in a distribution.
	Histogram(name string, value float64, tags map[string]string)

	// Timer starts a timer and returns a stop function. Calling the returned
	// function records the elapsed duration as a histogram observation.
	Timer(name string) func()
}
