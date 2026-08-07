// Package metrics exposes Prometheus HTTP metrics for the backend.
package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "simtas_http_requests_total",
			Help: "Total HTTP requests handled.",
		},
		[]string{"method", "status"},
	)
	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "simtas_http_request_duration_seconds",
			Help:    "HTTP request latency.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)

func init() {
	prometheus.MustRegister(httpRequests, httpDuration)
}

// Middleware records request counts and latency per method/status.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		httpRequests.WithLabelValues(c.Request.Method, statusLabel(c.Writer.Status())).Inc()
		httpDuration.WithLabelValues(c.Request.Method).Observe(time.Since(start).Seconds())
	}
}

// Handler serves the Prometheus scrape endpoint.
func Handler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

func statusLabel(code int) string {
	return string(rune('0'+code/100)) + "xx"
}
