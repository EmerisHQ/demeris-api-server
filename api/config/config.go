package config

import (
	"github.com/emerishq/emeris-utils/validation"

	"github.com/emerishq/emeris-utils/configuration"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	DatabaseConnectionURL  string `validate:"required"`
	ListenAddr             string `validate:"required"`
	RedisAddr              string `validate:"required"`
	KubernetesConfigMode   string
	KubernetesNamespace    string `validate:"required"`
	SentryDSN              string
	SentryEnvironment      string
	SentrySampleRate       float64
	SentryTracesSampleRate float64
	FeatureFlags           []string

	Debug bool
}

func (c Config) Validate() error {
	err := validator.New().Struct(c)
	if err != nil {
		return validation.MissingFieldsErr(err, false)
	}

	return nil
}

func Read() (*Config, error) {
	var c Config

	return &c, configuration.ReadConfig(&c, "demeris-api", map[string]string{
		"ListenAddr":             ":9090",
		"RedisAddr":              ":6379",
		"KubernetesNamespace":    "emeris",
		"SentryEnvironment":      "notset",
		"SentrySampleRate":       "1.0",
		"SentryTracesSampleRate": "0.01",
	})
}
