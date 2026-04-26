package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	DefaultAPIBase = "https://api-production-f26d.up.railway.app"
	DefaultIngest  = "https://api-production-f26d.up.railway.app"
)

type Config struct {
	APIKey  string `mapstructure:"api_key"`
	APIBase string `mapstructure:"api_base"`
	Ingest  string `mapstructure:"ingest_base"`
}

func Load() (*Config, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(dir)

	viper.SetDefault("api_base", DefaultAPIBase)
	viper.SetDefault("ingest_base", DefaultIngest)

	// Allow env overrides for local dev
	viper.SetEnvPrefix("GETHOOK")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
		// No config file yet — that's fine, defaults apply
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

func SaveAPIKey(key string) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	viper.Set("api_key", key)
	return viper.WriteConfigAs(filepath.Join(dir, "config.toml"))
}

func ClearAPIKey() error {
	viper.Set("api_key", "")
	dir, err := configDir()
	if err != nil {
		return err
	}
	return viper.WriteConfigAs(filepath.Join(dir, "config.toml"))
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home dir: %w", err)
	}
	return filepath.Join(home, ".config", "gethook"), nil
}
