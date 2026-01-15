package metrics

import (
	"net/http"

	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
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
