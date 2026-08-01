package config

import (
	"log/slog"
	"testing"
	"time"
)

func TestLoadFromLookupUsesDefaults(t *testing.T) {
	t.Parallel()

	config, err := LoadFromLookup(mapLookup(nil))
	if err != nil {
		t.Fatalf("LoadFromLookup() error = %v", err)
	}

	if config.Host != defaultHost {
		t.Errorf("Host = %q, want %q", config.Host, defaultHost)
	}
	if config.Port != defaultPort {
		t.Errorf("Port = %d, want %d", config.Port, defaultPort)
	}
	if config.ReadHeaderTimeout != defaultReadHeaderTimeout {
		t.Errorf(
			"ReadHeaderTimeout = %v, want %v",
			config.ReadHeaderTimeout,
			defaultReadHeaderTimeout,
		)
	}
	if config.ReadTimeout != defaultReadTimeout {
		t.Errorf(
			"ReadTimeout = %v, want %v",
			config.ReadTimeout,
			defaultReadTimeout,
		)
	}
	if config.WriteTimeout != defaultWriteTimeout {
		t.Errorf(
			"WriteTimeout = %v, want %v",
			config.WriteTimeout,
			defaultWriteTimeout,
		)
	}
	if config.IdleTimeout != defaultIdleTimeout {
		t.Errorf(
			"IdleTimeout = %v, want %v",
			config.IdleTimeout,
			defaultIdleTimeout,
		)
	}
	if config.ShutdownTimeout != defaultShutdownTimeout {
		t.Errorf(
			"ShutdownTimeout = %v, want %v",
			config.ShutdownTimeout,
			defaultShutdownTimeout,
		)
	}
	if config.MLServiceURL != defaultMLServiceURL {
		t.Errorf(
			"MLServiceURL = %q, want %q",
			config.MLServiceURL,
			defaultMLServiceURL,
		)
	}
	if config.MLServiceTimeout != defaultMLServiceTimeout {
		t.Errorf(
			"MLServiceTimeout = %v, want %v",
			config.MLServiceTimeout,
			defaultMLServiceTimeout,
		)
	}
	if config.LogLevel != defaultLogLevel {
		t.Errorf(
			"LogLevel = %v, want %v",
			config.LogLevel,
			defaultLogLevel,
		)
	}
}

func TestLoadFromLookupUsesOverrides(t *testing.T) {
	t.Parallel()

	values := map[string]string{
		"SEARCH_API_HOST":                "127.0.0.1",
		"SEARCH_API_PORT":                "9090",
		"SEARCH_API_READ_HEADER_TIMEOUT": "2s",
		"SEARCH_API_READ_TIMEOUT":        "3s",
		"SEARCH_API_WRITE_TIMEOUT":       "4s",
		"SEARCH_API_IDLE_TIMEOUT":        "5s",
		"SEARCH_API_SHUTDOWN_TIMEOUT":    "6s",
		"ML_SERVICE_URL":                 "http://127.0.0.1:9091/api/",
		"ML_SERVICE_TIMEOUT":             "750ms",
		"LOG_LEVEL":                      "DEBUG",
	}

	config, err := LoadFromLookup(mapLookup(values))
	if err != nil {
		t.Fatalf("LoadFromLookup() error = %v", err)
	}

	if config.Host != "127.0.0.1" {
		t.Errorf("Host = %q, want 127.0.0.1", config.Host)
	}
	if config.Port != 9090 {
		t.Errorf("Port = %d, want 9090", config.Port)
	}
	if config.ReadHeaderTimeout != 2*time.Second {
		t.Errorf(
			"ReadHeaderTimeout = %v, want 2s",
			config.ReadHeaderTimeout,
		)
	}
	if config.ReadTimeout != 3*time.Second {
		t.Errorf("ReadTimeout = %v, want 3s", config.ReadTimeout)
	}
	if config.WriteTimeout != 4*time.Second {
		t.Errorf("WriteTimeout = %v, want 4s", config.WriteTimeout)
	}
	if config.IdleTimeout != 5*time.Second {
		t.Errorf("IdleTimeout = %v, want 5s", config.IdleTimeout)
	}
	if config.ShutdownTimeout != 6*time.Second {
		t.Errorf(
			"ShutdownTimeout = %v, want 6s",
			config.ShutdownTimeout,
		)
	}
	if config.MLServiceURL != "http://127.0.0.1:9091/api" {
		t.Errorf(
			"MLServiceURL = %q, want http://127.0.0.1:9091/api",
			config.MLServiceURL,
		)
	}
	if config.MLServiceTimeout != 750*time.Millisecond {
		t.Errorf(
			"MLServiceTimeout = %v, want 750ms",
			config.MLServiceTimeout,
		)
	}
	if config.LogLevel != slog.LevelDebug {
		t.Errorf(
			"LogLevel = %v, want %v",
			config.LogLevel,
			slog.LevelDebug,
		)
	}
	if config.Address() != "127.0.0.1:9090" {
		t.Errorf(
			"Address() = %q, want 127.0.0.1:9090",
			config.Address(),
		)
	}
}

func TestAddressSupportsIPv6(t *testing.T) {
	t.Parallel()

	config := Config{Host: "::1", Port: 8080}
	if config.Address() != "[::1]:8080" {
		t.Errorf(
			"Address() = %q, want [::1]:8080",
			config.Address(),
		)
	}
}

func TestLoadFromLookupRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values map[string]string
	}{
		{
			name: "non-numeric port",
			values: map[string]string{
				"SEARCH_API_PORT": "not-a-number",
			},
		},
		{
			name: "port too large",
			values: map[string]string{
				"SEARCH_API_PORT": "70000",
			},
		},
		{
			name: "invalid duration",
			values: map[string]string{
				"SEARCH_API_READ_TIMEOUT": "soon",
			},
		},
		{
			name: "non-positive duration",
			values: map[string]string{
				"SEARCH_API_IDLE_TIMEOUT": "0s",
			},
		},
		{
			name: "invalid ML URL scheme",
			values: map[string]string{
				"ML_SERVICE_URL": "ftp://localhost:8090",
			},
		},
		{
			name: "ML URL without host",
			values: map[string]string{
				"ML_SERVICE_URL": "http:///missing-host",
			},
		},
		{
			name: "ML URL with credentials",
			values: map[string]string{
				"ML_SERVICE_URL": "http://user:pass@localhost:8090",
			},
		},
		{
			name: "ML URL with query",
			values: map[string]string{
				"ML_SERVICE_URL": "http://localhost:8090?debug=true",
			},
		},
		{
			name: "invalid ML timeout",
			values: map[string]string{
				"ML_SERVICE_TIMEOUT": "0s",
			},
		},
		{
			name: "invalid log level",
			values: map[string]string{
				"LOG_LEVEL": "verbose",
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if _, err := LoadFromLookup(
				mapLookup(test.values),
			); err == nil {
				t.Fatal("LoadFromLookup() error = nil, want error")
			}
		})
	}
}

func TestLoadFromLookupRejectsNilLookup(t *testing.T) {
	t.Parallel()

	if _, err := LoadFromLookup(nil); err == nil {
		t.Fatal("LoadFromLookup(nil) error = nil, want error")
	}
}

func mapLookup(values map[string]string) LookupEnv {
	return func(key string) (string, bool) {
		value, exists := values[key]
		return value, exists
	}
}
