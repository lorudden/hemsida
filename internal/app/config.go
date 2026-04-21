package app

import "context"

// Flag identifies an application startup flag.
type Flag int

// Flags stores resolved startup flag values.
type Flags map[Flag]string

const (
	ListenAddress Flag = iota
	ServicePort
	ControlPort
	WebAssetPath
	DevModeEnabled
	LogFormat
)

// DefaultFlags returns the baseline startup flags for the service.
func DefaultFlags() Flags {
	return Flags{
		ListenAddress:  "",
		ServicePort:    "8080",
		ControlPort:    "",
		WebAssetPath:   "/opt/lorudden/assets",
		DevModeEnabled: "false",
		LogFormat:      "json",
	}
}

// Config contains application configuration assembled during startup.
type Config struct{}

// NewConfig builds application configuration from the resolved flags.
func NewConfig(context.Context, Flags) (*Config, error) {
	return &Config{}, nil
}
