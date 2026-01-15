package metrics

import (
	"net/http"

	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	latencyBuckets = []float64{0.001, 0.01, 0.05, 0.1, 0.5, 1, 5, 10, 50, 100, 500, 1000}
	sizeBuckets    = []float64{1, 128, 1024, 128 * 1024, 1024 * 1024, 2 * 1024 * 1024, 8 * 1024 * 1024, 16 * 1024 * 1024, 128 * 1024 * 1024, 1024 * 1024 * 1024}

	HandlerNames = map[string]bool{
		"READ":             true,
		"WRITE":            true,
		"LIMIT_READ":       true,
		"LIMIT_WRITE":      true,
		"S3_PUT":           true,
		"S3_GET":           true,
		"CAT":              true,
		"CATV2":            true,
		"PUT":              true,
		"PUTV2":            true,
		"PUTV3":            true,
		"DELETE":           true,
		"LIST":             true,
		"LISTV2":           true,
		"OBJECT META":      true,
		"COPY":             true,
		"COPYV2":           true,
		"GOOL":             true,
		"ERROR":            true,
		"UNTRASHIFY":       true,
		"COLLECT OBSOLETE": true,
		"DELETE OBSOLETE":  true,
	}
)

func StoreLatencyAndSizeInfo(opType string, size float64, latency float64) {
	if _, ok := HandlerNames[opType]; !ok {
		return
	}
	HisstogramSizeVec.With(map[string]string{
		"source": opType,
	}).Observe(size)
	if size != 0 {
		HisstogramLatencyVec.With(map[string]string{
			"source": opType,
		}).Observe(latency / size)
	}
}

var (
	ReadReqProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "read_req_processed_total",
		Help: "The total number of processed reads",
	})
	ReadReqErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "read_req_errors_total",
		Help: "The total number of errors during reads",
	})
	WiteReqProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "write_req_processed_total",
		Help: "The total number of processed writes",
	})
	WriteReqErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "write_req_errors_total",
		Help: "The total number of errors during reads",
	})
	HisstogramLatencyVec = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "request_latency",
		Help:    "Request latency in seconds",
		Buckets: latencyBuckets,
	}, []string{"source"})

	HisstogramSizeVec = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "request_size",
		Help:    "Request latency in seconds",
		Buckets: sizeBuckets,
	}, []string{"source"})
)

func NewMetricsWebServer(httpAddr string) *MetricsWebServer {
	return &MetricsWebServer{
		httpAddr: httpAddr,
	}
}

type MetricsWebServer struct {
	*http.Server
	*http.ServeMux
	mu       sync.Mutex
	httpAddr string
}

func (mws *MetricsWebServer) Serve() error {
	mws.mu.Lock()
	defer mws.mu.Unlock()

	mws.configureDebugWebServer()

	go func() {
		_ = mws.ListenAndServe()
	}()

	return nil
}

func (mws *MetricsWebServer) configureDebugWebServer() {
	mux := http.NewServeMux()
	mws.Server = &http.Server{
		Addr:    mws.httpAddr,
		Handler: mux,
	}
	mws.ServeMux = mux

	mws.enablePrometheusEndpoints()
}

// enablePrometheusEndpoints exposes prometheus http endpoints.
func (dws *MetricsWebServer) enablePrometheusEndpoints() {
	dws.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)
}
