package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusHttpTraffic interface {
	RequestTotalCollector() *prometheus.CounterVec
	RequestDurationCollector() *prometheus.HistogramVec
	Register(collector []prometheus.Collector) error
}

type prometheusHttpTrafficImpl struct {
	prometheusRegistry *prometheus.Registry
}

func NewPrometheusHttpTraffic(prometheusRegistry *prometheus.Registry) PrometheusHttpTraffic {
	return &prometheusHttpTrafficImpl{
		prometheusRegistry: prometheusRegistry,
	}
}

func (m *prometheusHttpTrafficImpl) RequestTotalCollector() *prometheus.CounterVec {
	return prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_request_total",
		Help: "Total number of HTTP requests",
	}, []string{"status_code"})
}

func (m *prometheusHttpTrafficImpl) RequestDurationCollector() *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_second",
		Help:    "HTTP request duration in second",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status_code"})
}

func (m *prometheusHttpTrafficImpl) Register(collector []prometheus.Collector) error {
	for _, c := range collector {
		if err := m.prometheusRegistry.Register(c); err != nil {
			return err
		}
	}
	return nil
}

func RequestDuration(reqTotalCollector *prometheus.CounterVec, reqDurationCollector *prometheus.HistogramVec) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			return
		}
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()

		responseStatus := c.Writer.Status()
		reqPath := c.FullPath()
		httpMethod := c.Request.Method
		reqTotalCollector.WithLabelValues(strconv.Itoa(responseStatus)).Inc()
		reqDurationCollector.WithLabelValues(httpMethod, reqPath, strconv.Itoa(responseStatus)).Observe(duration)
	}
}
