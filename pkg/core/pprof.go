package core

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"

	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

func NewDebugWebServer(httpAddr string) *DebugWebServer {
	return &DebugWebServer{
		httpAddr: httpAddr,
	}
}

type DebugWebServer struct {
	*http.Server
	*http.ServeMux
	mu       sync.Mutex
	httpAddr string
	timer    *time.Timer
}

// ServeFor starts pprof webserver in the background goroutine.
// After Duration `t` server shuts down for security reasons.
func (dws *DebugWebServer) ServeFor(t time.Duration) error {
	dws.mu.Lock()
	defer dws.mu.Unlock()

	if dws.timer != nil {
		postponed := dws.timer.Reset(t)
		if !postponed {
			// Reset() started new countdown (because previous were executed), however we just
			// observed that `dws.timer` is not nil... it means that AfterFunc`s is running
			// concurrently (waiting for mutex)
			return fmt.Errorf("previous debug server is being shutdown, retry later")
		}
		return nil
	}

	dws.configureDebugWebServer()
	dws.timer = time.AfterFunc(t, func() {
		err := dws.Shutdown(context.Background())
		if err != nil {
			ylogger.Zero.Error().Err(err).Msg("Failed to shutdown debug webserver")
		}
	})
	go func() {
		_ = dws.ListenAndServe()
	}()

	return nil
}

func (dws *DebugWebServer) configureDebugWebServer() {
	mux := http.NewServeMux()
	dws.Server = &http.Server{
		Addr:    dws.httpAddr,
		Handler: mux,
	}
	dws.ServeMux = mux

	dws.enablePprofEndpoints()
	dws.enableExpVarEndpoints()
}

// Shutdown synchronously shuts down debug server (if started).
func (dws *DebugWebServer) Shutdown(ctx context.Context) error {
	dws.mu.Lock()
	defer dws.mu.Unlock()

	if dws.timer == nil {
		return nil
	}
	dws.timer = nil
	return dws.Server.Shutdown(ctx)
}

// enablePprofEndpoints exposes pprof http endpoints.
func (dws *DebugWebServer) enablePprofEndpoints() {
	dws.HandleFunc("/debug/pprof/", pprof.Index)
	dws.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	dws.HandleFunc("/debug/pprof/profile", pprof.Profile)
	dws.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	dws.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

// enableExpVarEndpoints exposes expvar http endpoints.
func (dws *DebugWebServer) enableExpVarEndpoints() {
	dws.HandleFunc("/debug/vars", expvar.Handler().ServeHTTP)
}
