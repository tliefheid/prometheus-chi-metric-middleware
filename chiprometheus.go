package chiprometheus

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

type Options struct {
	DisableRequestCounter   bool
	DisableRequestDurations bool
	DisableResponseSize     bool
	Namespace               string
	Subsystem               string
	ConstLabels             map[string]string
	Debug                   bool
}

type Instance struct {
	reqCount                *prometheus.CounterVec
	reqDuration             *prometheus.HistogramVec
	respSize                *prometheus.HistogramVec
	disableRequestCounter   bool
	disableRequestDurations bool
	disableResponseSize     bool
	debug                   bool
}

func NewMiddleware(opt Options) *Instance {
	i := new(Instance)
	i.disableRequestCounter = opt.DisableRequestCounter
	i.disableRequestDurations = opt.DisableRequestDurations
	i.disableResponseSize = opt.DisableResponseSize
	i.debug = opt.Debug

	if !i.disableRequestCounter {
		i.reqCount = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   opt.Namespace,
				Subsystem:   opt.Subsystem,
				Name:        "http_request_total",
				Help:        "Counter of HTTP Requests",
				ConstLabels: opt.ConstLabels,
			}, []string{"code", "path"},
		)
		prometheus.MustRegister(i.reqCount)
	}

	if !i.disableRequestDurations {
		i.reqDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   opt.Namespace,
				Subsystem:   opt.Subsystem,
				Name:        "http_request_duration_nanoseconds",
				Help:        "Histogram of latencies for HTTP requests.",
				Buckets:     []float64{.1, .2, .4, 1, 3, 8, 20, 60, 120},
				ConstLabels: opt.ConstLabels,
			},
			[]string{"path"})
		prometheus.MustRegister(i.reqDuration)
	}

	if !i.disableResponseSize {
		i.respSize = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   opt.Namespace,
				Subsystem:   opt.Subsystem,
				Name:        "http_response_size_bytes",
				Help:        "Histogram of response size for HTTP requests.",
				Buckets:     prometheus.ExponentialBuckets(100, 10, 8),
				ConstLabels: opt.ConstLabels,
			},
			[]string{"path"})
		prometheus.MustRegister(i.respSize)
	}
	return i
}

func (i *Instance) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrap := middleware.NewWrapResponseWriter(rw, r.ProtoMajor)
		next.ServeHTTP(wrap, r)

		rctx := chi.RouteContext(r.Context())
		routePattern := strings.Join(rctx.RoutePatterns, "")
		path := strings.ReplaceAll(routePattern, "//", "/")
		path = strings.Replace(path, "/*/", "/", -1)
		if i.debug {
			fmt.Printf("Handle metrics function\nRoutePattern: %+v\nPath: %v\nStatusCode: %v\n", routePattern, path, wrap.Status())
		}
		if !i.disableRequestCounter {
			i.reqCount.WithLabelValues(strconv.Itoa(wrap.Status()), path).Inc()
		}
		if !i.disableRequestDurations {
			i.reqDuration.WithLabelValues(path).Observe(float64(time.Since(start).Nanoseconds()))
		}
		if !i.disableResponseSize {
			i.respSize.WithLabelValues(path).Observe(float64(wrap.BytesWritten()))
		}
	})
}
