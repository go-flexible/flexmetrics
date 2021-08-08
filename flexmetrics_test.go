package flexmetrics_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/go-flexible/flexmetrics"
)

var (
	ctx context.Context
)

func ExampleServer_Run() {
	srv := flexmetrics.New(&flexmetrics.Config{
		Server: &http.Server{
			Addr: "0.0.0.0:5117",
		},
		Path: "/metrics",
	})
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
		config           *flexmetrics.Config
		expectedAddress  string
		expectedEndpoint string
	}{
		{
			name: "environment sets address and path",
			env: map[string]string{
				"METRICS_ADDR":            "0.0.0.0:1111",
				"METRICS_PROMETHEUS_PATH": "/testmetrics",
			},
			config:           nil,
			expectedAddress:  "0.0.0.0:1111",
			expectedEndpoint: "/testmetrics",
		},
		{
			name:             "default config doesn't override environment",
			config:           &flexmetrics.Config{},
			expectedAddress:  "0.0.0.0:1111",
			expectedEndpoint: "/testmetrics",
		},
		{
			name: "address is overridden in config",
			config: &flexmetrics.Config{
				Server: &http.Server{Addr: "0.0.0.0:2222"},
			},
			expectedAddress:  "0.0.0.0:2222",
			expectedEndpoint: "/testmetrics",
		},
		{
			name: "endpoint is overridden in config but address is from env",
			config: &flexmetrics.Config{
				Path: "/zero",
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
			srv := flexmetrics.New(tt.config)
			if tt.expectedAddress != srv.Server.Addr {
				t.Errorf("%s: expected address %q, but got %q", tt.name, tt.expectedAddress, srv.Server.Addr)
			}
			if tt.expectedEndpoint != srv.Path {
				t.Errorf("%s: expected endpoint %q, but got %q", tt.name, tt.expectedEndpoint, srv.Path)
			}
		})
	}
}
