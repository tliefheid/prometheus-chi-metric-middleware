package chiprometheus

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

type Options struct {
	disableRequestCounter   bool
	disableRequestDurations bool
	disableResponseSize     bool
	Namespace               string
	Subsystem               string
	ConstLabels             map[string]string
}

type Instance struct {
	reqCount                *prometheus.CounterVec
	reqDuration             *prometheus.HistogramVec
	respSize                *prometheus.HistogramVec
	disableRequestCounter   bool
	disableRequestDurations bool
	disableResponseSize     bool
}

func NewMiddleware(opt Options) *Instance {
	i := new(Instance)
	i.disableRequestCounter = opt.disableRequestCounter
	i.disableRequestDurations = opt.disableRequestDurations
	i.disableResponseSize = opt.disableResponseSize

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
		ww := middleware.NewWrapResponseWriter(rw, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		rctx := chi.RouteContext(r.Context())
		routePattern := strings.Join(rctx.RoutePatterns, "")
		path := strings.Replace(routePattern, "/*/", "/", -1)

		i.reqCount.WithLabelValues(strconv.Itoa(ww.Status()), path).Inc()
		i.reqDuration.WithLabelValues(path).Observe(float64(time.Since(start).Nanoseconds()))
		i.respSize.WithLabelValues(path).Observe(float64(ww.BytesWritten()))
	})
}
