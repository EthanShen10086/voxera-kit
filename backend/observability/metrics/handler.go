package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HTTPHandler returns an [http.Handler] that serves Prometheus metrics in the
// standard exposition format. Mount it on your HTTP mux (typically at
// "/metrics") to expose collected metrics to a Prometheus scraper.
func HTTPHandler() http.Handler {
	return promhttp.Handler()
}
