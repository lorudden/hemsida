package app

import (
	"context"
	"flag"

	"github.com/diwise/service-chassis/pkg/infrastructure/env"
)

// Flag identifies an application startup flag.
type Flag int

// Flags stores resolved startup flag values.
type Flags map[Flag]string

const (
	ListenAddress Flag = iota
	ServicePort
	ControlPort
	WebAssetPath
	MediaDataPath
	ContentDataPath
	DevModeEnabled
	LogFormat
)

// DefaultFlags returns the baseline startup flags for the service.
func DefaultFlags() Flags {
	return Flags{
		ListenAddress:   "",
		ServicePort:     "8080",
		ControlPort:     "",
		WebAssetPath:    "/opt/lorudden/assets",
		MediaDataPath:   "data/media",
		ContentDataPath: "data/content",
		DevModeEnabled:  "false",
		LogFormat:       "json",
	}
}

// ResolveFlags merges defaults, environment variables, and command-line flags.
func ResolveFlags(ctx context.Context) (context.Context, Flags) {
	flags := DefaultFlags()

	flags[ControlPort] = env.GetVariableOrDefault(ctx, "CONTROL_PORT", flags[ControlPort])
	flags[ServicePort] = env.GetVariableOrDefault(ctx, "SERVICE_PORT", flags[ServicePort])
	flags[MediaDataPath] = env.GetVariableOrDefault(ctx, "MEDIA_DATA_PATH", flags[MediaDataPath])
	flags[ContentDataPath] = env.GetVariableOrDefault(ctx, "CONTENT_DATA_PATH", flags[ContentDataPath])

	apply := func(f Flag) func(string) error {
		return func(value string) error {
			flags[f] = value
			return nil
		}
	}

	// Allow command line arguments to override defaults and environment variables.
	flag.BoolFunc("devmode", "enable devmode with fake backend data", apply(DevModeEnabled))
	flag.Func("listen", "network and address to listen to", apply(ListenAddress))
	flag.Func("controlport", "port number to bind to for the control interface", apply(ControlPort))
	flag.Func("port", "port number to bind to for the public interface", apply(ServicePort))
	flag.Func("web-assets", "path to web assets folder", apply(WebAssetPath))
	flag.Func("media-data", "path to the JSON media catalog directory", apply(MediaDataPath))
	flag.Func("content-data", "path to the JSON content catalog directory", apply(ContentDataPath))
	flag.Func("log-format", "choose to get log output in text or json format", apply(LogFormat))
	flag.Parse()

	return ctx, flags
}

// Config contains application configuration assembled during startup.
type Config struct {
	MediaDataPath   string
	ContentDataPath string
}

// NewConfig builds application configuration from the resolved flags.
func NewConfig(_ context.Context, flags Flags) (*Config, error) {
	return &Config{
		MediaDataPath:   flags[MediaDataPath],
		ContentDataPath: flags[ContentDataPath],
	}, nil
}
