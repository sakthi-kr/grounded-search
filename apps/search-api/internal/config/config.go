package config

import (
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHost              = "0.0.0.0"
	defaultPort              = 8080
	defaultReadHeaderTimeout = 5 * time.Second
	defaultReadTimeout       = 10 * time.Second
	defaultWriteTimeout      = 15 * time.Second
	defaultIdleTimeout       = 60 * time.Second
	defaultShutdownTimeout   = 10 * time.Second
	defaultMLServiceURL      = "http://localhost:8090"
	defaultMLServiceTimeout  = 2 * time.Second
	defaultLogLevel          = slog.LevelInfo
)

// Config contains the runtime configuration for the search API.
type Config struct {
	Host              string
	Port              int
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MLServiceURL      string
	MLServiceTimeout  time.Duration
	LogLevel          slog.Level
}

// LookupEnv matches os.LookupEnv and makes configuration tests deterministic.
type LookupEnv func(string) (string, bool)

// Load reads configuration from the current process environment.
func Load() (Config, error) {
	return LoadFromLookup(os.LookupEnv)
}

// LoadFromLookup reads and validates configuration using the supplied lookup
// function.
func LoadFromLookup(lookup LookupEnv) (Config, error) {
	if lookup == nil {
		return Config{}, fmt.Errorf("environment lookup function is nil")
	}

	host := stringValue(lookup, "SEARCH_API_HOST", defaultHost)

	port, err := intValue(
		lookup,
		"SEARCH_API_PORT",
		defaultPort,
		1,
		65535,
	)
	if err != nil {
		return Config{}, err
	}

	readHeaderTimeout, err := durationValue(
		lookup,
		"SEARCH_API_READ_HEADER_TIMEOUT",
		defaultReadHeaderTimeout,
	)
	if err != nil {
		return Config{}, err
	}

	readTimeout, err := durationValue(
		lookup,
		"SEARCH_API_READ_TIMEOUT",
		defaultReadTimeout,
	)
	if err != nil {
		return Config{}, err
	}

	writeTimeout, err := durationValue(
		lookup,
		"SEARCH_API_WRITE_TIMEOUT",
		defaultWriteTimeout,
	)
	if err != nil {
		return Config{}, err
	}

	idleTimeout, err := durationValue(
		lookup,
		"SEARCH_API_IDLE_TIMEOUT",
		defaultIdleTimeout,
	)
	if err != nil {
		return Config{}, err
	}

	shutdownTimeout, err := durationValue(
		lookup,
		"SEARCH_API_SHUTDOWN_TIMEOUT",
		defaultShutdownTimeout,
	)
	if err != nil {
		return Config{}, err
	}

	mlServiceURL, err := urlValue(
		lookup,
		"ML_SERVICE_URL",
		defaultMLServiceURL,
	)
	if err != nil {
		return Config{}, err
	}

	mlServiceTimeout, err := durationValue(
		lookup,
		"ML_SERVICE_TIMEOUT",
		defaultMLServiceTimeout,
	)
	if err != nil {
		return Config{}, err
	}

	logLevel, err := logLevelValue(
		lookup,
		"LOG_LEVEL",
		defaultLogLevel,
	)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Host:              host,
		Port:              port,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ShutdownTimeout:   shutdownTimeout,
		MLServiceURL:      mlServiceURL,
		MLServiceTimeout:  mlServiceTimeout,
		LogLevel:          logLevel,
	}, nil
}

// Address returns a host:port value suitable for http.Server.
func (config Config) Address() string {
	return net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
}

func stringValue(
	lookup LookupEnv,
	key string,
	fallback string,
) string {
	value, exists := lookup(key)
	if !exists {
		return fallback
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}

	return value
}

func intValue(
	lookup LookupEnv,
	key string,
	fallback int,
	minimum int,
	maximum int,
) (int, error) {
	raw, exists := lookup(key)
	if !exists || strings.TrimSpace(raw) == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}

	if value < minimum || value > maximum {
		return 0, fmt.Errorf(
			"%s must be between %d and %d",
			key,
			minimum,
			maximum,
		)
	}

	return value, nil
}

func durationValue(
	lookup LookupEnv,
	key string,
	fallback time.Duration,
) (time.Duration, error) {
	raw, exists := lookup(key)
	if !exists || strings.TrimSpace(raw) == "" {
		return fallback, nil
	}

	value, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf(
			"%s must be a valid Go duration: %w",
			key,
			err,
		)
	}

	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}

	return value, nil
}

func urlValue(
	lookup LookupEnv,
	key string,
	fallback string,
) (string, error) {
	raw := stringValue(lookup, key, fallback)
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("%s must be a valid URL: %w", key, err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf(
			"%s scheme must be http or https",
			key,
		)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("%s must include a host", key)
	}
	if parsed.User != nil {
		return "", fmt.Errorf("%s must not include credentials", key)
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf(
			"%s must not include a query or fragment",
			key,
		)
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/")

	return parsed.String(), nil
}

func logLevelValue(
	lookup LookupEnv,
	key string,
	fallback slog.Level,
) (slog.Level, error) {
	raw, exists := lookup(key)
	if !exists || strings.TrimSpace(raw) == "" {
		return fallback, nil
	}

	var level slog.Level
	if err := level.UnmarshalText(
		[]byte(strings.TrimSpace(raw)),
	); err != nil {
		return 0, fmt.Errorf("%s is invalid: %w", key, err)
	}

	return level, nil
}
