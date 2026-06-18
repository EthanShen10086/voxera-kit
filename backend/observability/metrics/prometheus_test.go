package metrics_test

import (
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/observability/metrics"
)

func TestPrometheusRecorder(t *testing.T) {
	r := metrics.NewPrometheusRecorder()
	r.Counter("kit_test_counter_total", 1, map[string]string{"env": "test"})
	r.Gauge("kit_test_gauge", 42, map[string]string{"env": "test"})
	r.Histogram("kit_test_hist", 0.5, map[string]string{"env": "test"})
	stop := r.Timer("kit_test_timer_seconds")
	time.Sleep(time.Millisecond)
	stop()
}

func TestHTTPHandler(t *testing.T) {
	h := metrics.HTTPHandler()
	if h == nil {
		t.Fatal("nil handler")
	}
}
