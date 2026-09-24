package middleware

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "townhall_http_requests_total",
			Help: "Total HTTP requests by method, route, and status.",
		},
		[]string{"method", "route", "status"},
	)
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "townhall_http_request_duration_seconds",
			Help:    "HTTP request latency by method and route.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)
)

// RegisterMetrics adds instrumentation plus the /metrics endpoint itself.
func RegisterMetrics(e *echo.Echo) {
	e.Use(Metrics())
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
}

// Metrics records count and latency per method/route/status.
func Metrics() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if c.Path() == "/metrics" {
				return next(c)
			}
			start := time.Now()
			err := next(c)
			route := c.Path()
			if route == "" {
				route = "unknown"
			}
			_, status := echo.ResolveResponseStatus(c.Response(), err)
			httpRequestsTotal.WithLabelValues(c.Request().Method, route, strconv.Itoa(status)).Inc()
			httpRequestDuration.WithLabelValues(c.Request().Method, route).Observe(time.Since(start).Seconds())
			return err
		}
	}
}
