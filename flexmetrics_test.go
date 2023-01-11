package flexmetrics_test

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

	"github.com/go-flexible/flexmetrics"
)

var ctx context.Context

func ExampleServer_Run() {
	srv := flexmetrics.New()
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
		options          []flexmetrics.Option
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
			options: []flexmetrics.Option{
				flexmetrics.WithAddr("0.0.0.0:2222"),
			},
			expectedAddress:  "0.0.0.0:2222",
			expectedEndpoint: "/testmetrics",
		},
		{
			name: "endpoint is overridden in config but address is from env",
			options: []flexmetrics.Option{
				flexmetrics.WithPath("/zero"),
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
			srv := flexmetrics.New(tt.options...)
			if tt.expectedAddress != srv.Server.Addr {
				t.Errorf("%s: expected address %q, but got %q", tt.name, tt.expectedAddress, srv.Server.Addr)
			}
			if tt.expectedEndpoint != srv.Path {
				t.Errorf("%s: expected endpoint %q, but got %q", tt.name, tt.expectedEndpoint, srv.Path)
			}
		})
	}
}

func TestOption_WithServer(t *testing.T) {
	myServer := &http.Server{ReadHeaderTimeout: time.Second}
	s := flexmetrics.New(flexmetrics.WithServer(myServer))
	if s.Server != myServer {
		t.Error("WithServer option should set the provided http server")
	}
}

func TestOption_WithLogger(t *testing.T) {
	var buf bytes.Buffer

	w := io.MultiWriter(&buf, os.Stderr)     // so we get console output.
	logger := log.New(w, "TEST_LOGGER: ", 0) // so we get consistent output.

	metrics := flexmetrics.New(
		flexmetrics.WithAddr("127.0.0.1:"),
		flexmetrics.WithLogger(logger),
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
