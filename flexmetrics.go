package flexmetrics

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// DefaultAddr is the port that we listen to the prometheus path on by default.
	DefaultAddr = "0.0.0.0:9090"

	// DefaultPath is the path where we expose prometheus by default.
	DefaultPath = "/metrics"

	// DefaultReadTimeout is the default read timeout for the http server.
	DefaultReadTimeout = 5 * time.Second

	// DefaultReadHeaderTimeout is the default read header timeout for the http server.
	DefaultReadHeaderTimeout = 1 * time.Second

	// DefaultIdleTimeout is the default idle timeout for the http server.
	DefaultIdleTimeout = 1 * time.Second

	// DefaultWriteTimeout is the default write timeout for the http server.
	DefaultWriteTimeout = 15 * time.Second
)

// default package logger to be used if no logger is provided.
var logger = log.New(os.Stderr, "flexmetrics: ", 0)

// Option is a type of func that allows you change defaults of the *Server
// returned by New.
type Option func(s *Server)

// WithPath allows you to specify the path Prometheus metrics should be served
// on. If provided, this option will override the use of
// METRICS_PROMETHEUS_PATH environment variable.
func WithPath(path string) Option {
	return func(s *Server) {
		s.Path = path
	}
}

// WithAddr allows you to specify the address the HTTP server should bind to.
// If provided, this option will override the use of METRICS_ADDR environment
// variable.
func WithAddr(addr string) Option {
	return func(s *Server) {
		s.Server.Addr = addr
	}
}

// WithServer allows you to provide your own HTTP server to be used to serve
// metrics.
func WithServer(server *http.Server) Option {
	return func(s *Server) {
		s.Server = server
	}
}

// Logger defines any logger able to call Printf.
type Logger interface {
	Printf(format string, v ...interface{})
}

// WithLogger allows you to set a logger for the server.
func WithLogger(l Logger) Option {
	return func(s *Server) {
		s.logger = l
	}
}

// New creates a new default metrics server.
func New(options ...Option) *Server {
	path := os.Getenv("METRICS_PROMETHEUS_PATH")
	if path == "" {
		path = DefaultPath
	}

	addr := os.Getenv("METRICS_ADDR")
	if addr == "" {
		addr = DefaultAddr
	}

	server := &Server{
		logger: logger,
		Path:   path,
		Server: &http.Server{
			Addr:              addr,
			ReadTimeout:       DefaultReadTimeout,
			ReadHeaderTimeout: DefaultReadHeaderTimeout,
			IdleTimeout:       DefaultIdleTimeout,
			WriteTimeout:      DefaultWriteTimeout,
		},
	}

	for _, option := range options {
		option(server)
	}

	return server
}

// Server represents a prometheus metrics server.
type Server struct {
	logger Logger
	Server *http.Server
	Path   string
}

// serverAddrKeyType is a type used to store the server address in the context.
type serverAddrKeyType string

// serverAddrKey is the key used to store the server address in the context.
const serverAddrKey serverAddrKeyType = "serverAddr"

// Run will start the metrics server.
func (s *Server) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.Server.Addr)
	if err != nil {
		return err
	}

	// use the provided context for the server
	s.Server.BaseContext = func(lis net.Listener) context.Context {
		ctx = context.WithValue(ctx, serverAddrKey, lis.Addr().String())
		return ctx
	}

	mux := http.NewServeMux()
	mux.Handle(s.Path, promhttp.Handler())
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	s.Server.Handler = mux
	s.logger.Printf("serving profiling and prometheus metrics over http on http://%s%s", lis.Addr().String(), s.Path)
	return s.Server.Serve(lis)
}

// Halt will attempt to gracefully shut down the server.
func (s *Server) Halt(ctx context.Context) error {
	listenerAddress, ok := ctx.Value(serverAddrKey).(string)
	if !ok {
		return fmt.Errorf("listener address not found in context")
	}
	s.logger.Printf("stopping serving profiling and prometheus metrics over http on http://%s%s", listenerAddress, s.Path)
	return s.Server.Shutdown(ctx)
}
