package config

import (
	"fmt"

	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Server struct {
		Port int
	}
	GitHub struct {
		AppID          int64
		InstallationID int64
		WebhookSecret  string
		PrivateKeyPath string
	}
	Backups struct {
		Strategy  string
		Bucket    string
		Directory string
	}
}

func Load(logger *zap.Logger) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("backup.directory", "./backups")

	viper.SetEnvPrefix("GITSAVER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Warn("no config file found, using defaults and environment variables")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
