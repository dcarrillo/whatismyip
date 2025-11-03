package metrics

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	enabled          bool
	initOnce         sync.Once
	requestsTotal    *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	requestsInFlight prometheus.Gauge
	geoLookups       *prometheus.CounterVec
	portScans        prometheus.Counter
	dnsQueries       *prometheus.CounterVec
)

func Enable() {
	initOnce.Do(func() {
		enabled = true

		requestsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "whatismyip_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		)

		requestDuration = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "whatismyip_http_request_duration_seconds",
				Help:    "HTTP request latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		)

		requestsInFlight = promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "whatismyip_http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
		)

		geoLookups = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "whatismyip_geo_lookups_total",
				Help: "Total number of geo lookups",
			},
			[]string{"type"},
		)

		portScans = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "whatismyip_port_scans_total",
				Help: "Total number of port scan requests",
			},
		)

		dnsQueries = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "whatismyip_dns_queries_total",
				Help: "Total number of DNS queries",
			},
			[]string{"query_type", "rcode"},
		)
	})
}

func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enabled {
			c.Next()
			return
		}

		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "/404" // group 404s
		}

		requestsInFlight.Inc()
		defer requestsInFlight.Dec()

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		requestsTotal.WithLabelValues(c.Request.Method, path, fmt.Sprintf("%dxx", status/100)).Inc()
		requestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

func RecordGeoLookup(lookupType string) {
	if !enabled {
		return
	}
	geoLookups.WithLabelValues(lookupType).Inc()
}

func RecordPortScan() {
	if !enabled {
		return
	}
	portScans.Inc()
}

func RecordDNSQuery(queryType string, rcode string) {
	if !enabled {
		return
	}
	dnsQueries.WithLabelValues(queryType, rcode).Inc()
}
