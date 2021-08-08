package flexmetrics

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"path"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// DefaultAddr is the port that we listen to the prometheus path on by default.
	DefaultAddr = "0.0.0.0:2112"

	// DefaultPath is the path where we expose prometheus by default.
	DefaultPath = "/metrics"
)

// Config represents the configuration for the metrics server.
type Config struct {
	Path   string
	Server *http.Server
}

// New creates a new default metrics server.
func New(config *Config) *Server {
	if config == nil {
		config = &Config{}
	}
	if prometheusPath := os.Getenv("METRICS_PROMETHEUS_PATH"); prometheusPath != "" && config.Path == "" {
		config.Path = prometheusPath
	}
	if config.Path == "" {
		config.Path = DefaultPath
	}
	if config.Server == nil {
		config.Server = &http.Server{}
	}
	if addr := os.Getenv("METRICS_ADDR"); addr != "" && config.Server.Addr == "" {
		config.Server.Addr = addr
	}
	if config.Server.Addr == "" {
		config.Server.Addr = DefaultAddr
	}
	return &Server{
		Path:   path.Join("/", config.Path),
		Server: config.Server,
	}
}

// Server represents a prometheus metrics server.
type Server struct {
	Path   string
	Server *http.Server
}

// Run will start the metrics server.
func (s *Server) Run(_ context.Context) error {
	lis, err := net.Listen("tcp", s.Server.Addr)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle(s.Path, promhttp.Handler())
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	s.Server.Handler = mux
	log.Printf("serving profiling and prometheus metrics over http on http://%s%s", s.Server.Addr, s.Path)
	return s.Server.Serve(lis)
}

// Halt will attempt to gracefully shut down the server.
func (s *Server) Halt(ctx context.Context) error {
	log.Printf("stopping serving profiling and prometheus metrics over http on http://%s...", s.Server.Addr)
	return s.Server.Shutdown(ctx)
}
