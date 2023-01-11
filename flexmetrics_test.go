package flexmetrics

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

var ctx context.Context

func ExampleServer_Run() {
	srv := New()
	_ = srv.Run(ctx)
}

func TestNew(t *testing.T) {
	t.Cleanup(func() {
		os.Unsetenv("METRICS_ADDR")
		os.Unsetenv("METRICS_PROMETHEUS_PATH")
	})
	cases := []struct {
		name             string
		env              map[string]string
		options          []Option
		expectedAddress  string
		expectedEndpoint string
	}{
		{
			name: "environment sets address and path",
			env: map[string]string{
				"METRICS_ADDR":            "0.0.0.0:1111",
				"METRICS_PROMETHEUS_PATH": "/testmetrics",
			},
			options:          nil,
			expectedAddress:  "0.0.0.0:1111",
			expectedEndpoint: "/testmetrics",
		},
		{
			name:             "no options provided overriding environment",
			options:          nil,
			expectedAddress:  "0.0.0.0:1111",
			expectedEndpoint: "/testmetrics",
		},
		{
			name: "address is overridden with option",
			options: []Option{
				WithAddr("0.0.0.0:2222"),
			},
			expectedAddress:  "0.0.0.0:2222",
			expectedEndpoint: "/testmetrics",
		},
		{
			name: "endpoint is overridden in config but address is from env",
			options: []Option{
				WithPath("/zero"),
			},
			expectedAddress:  "0.0.0.0:1111",
			expectedEndpoint: "/zero",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.env {
				os.Setenv(key, val)
			}
			srv := New(tt.options...)
			if tt.expectedAddress != srv.server.Addr {
				t.Errorf("%s: expected address %q, but got %q", tt.name, tt.expectedAddress, srv.server.Addr)
			}
			if tt.expectedEndpoint != srv.path {
				t.Errorf("%s: expected endpoint %q, but got %q", tt.name, tt.expectedEndpoint, srv.path)
			}
		})
	}
}

func TestOption_WithServer(t *testing.T) {
	myServer := &http.Server{ReadHeaderTimeout: time.Second}
	s := New(WithServer(myServer))
	if s.server != myServer {
		t.Error("WithServer option should set the provided http server")
	}
}

func TestOption_WithLogger(t *testing.T) {
	var buf bytes.Buffer

	w := io.MultiWriter(&buf, os.Stderr)     // so we get console output.
	logger := log.New(w, "TEST_LOGGER: ", 0) // so we get consistent output.

	metrics := New(
		WithAddr("127.0.0.1:"),
		WithLogger(logger),
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	go func() {
		_ = metrics.Run(ctx)
	}()
	time.Sleep(time.Second)
	_ = metrics.Halt(ctx)

	t.Log(buf.String())

	// ugly? yes, but, it will do.
	if !strings.Contains(buf.String(), "TEST_LOGGER: ") {
		t.Fatal("expected log message to contain prefix")
	}
}
