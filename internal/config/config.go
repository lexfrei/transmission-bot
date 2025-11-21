// Package config provides configuration loading and validation for the transmission-bot.
package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Validation errors.
var (
	ErrMissingToken        = errors.New("telegram.token is required")
	ErrMissingAllowedUsers = errors.New("telegram.allowed_users is required (at least one user ID)")
	ErrMissingURL          = errors.New("transmission.url is required")
)

// Config holds all configuration for the application.
type Config struct {
	Telegram     TelegramConfig     `mapstructure:"telegram"`
	Transmission TransmissionConfig `mapstructure:"transmission"`
	Log          LogConfig          `mapstructure:"log"`
}

// TelegramConfig holds Telegram bot configuration.
type TelegramConfig struct {
	Token        string  `mapstructure:"token"`
	AllowedUsers []int64 `mapstructure:"allowed_users"`
}

// TransmissionConfig holds Transmission RPC configuration.
type TransmissionConfig struct {
	URL      string `mapstructure:"url"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string `mapstructure:"level"`
}

// Load reads configuration from file, environment variables, and defaults.
func Load(configPath string) (*Config, error) {
	viperInstance := viper.New()

	viperInstance.SetDefault("transmission.url", "http://localhost:9091/transmission/rpc")
	viperInstance.SetDefault("log.level", "info")

	viperInstance.SetEnvPrefix("TB")
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viperInstance.AutomaticEnv()

	if configPath != "" {
		viperInstance.SetConfigFile(configPath)
	} else {
		viperInstance.SetConfigName("config")
		viperInstance.SetConfigType("yaml")
		viperInstance.AddConfigPath(".")
		viperInstance.AddConfigPath("$HOME/.transmission-bot")
		viperInstance.AddConfigPath("/etc/transmission-bot")
	}

	readErr := viperInstance.ReadInConfig()
	if readErr != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(readErr, &configFileNotFoundError) {
			return nil, fmt.Errorf("reading config: %w", readErr)
		}
	}

	var cfg Config

	unmarshalErr := viperInstance.Unmarshal(&cfg)
	if unmarshalErr != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", unmarshalErr)
	}

	validateErr := cfg.Validate()
	if validateErr != nil {
		return nil, fmt.Errorf("validating config: %w", validateErr)
	}

	return &cfg, nil
}

// Validate checks that all required configuration fields are set.
func (c *Config) Validate() error {
	if c.Telegram.Token == "" {
		return ErrMissingToken
	}

	if len(c.Telegram.AllowedUsers) == 0 {
		return ErrMissingAllowedUsers
	}

	if c.Transmission.URL == "" {
		return ErrMissingURL
	}

	return nil
}
