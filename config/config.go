package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	DBURL       string `mapstructure:"DB_URL"`
	DBToken     string `mapstructure:"DB_TOKEN"`
	DBMode      string `mapstructure:"DB_MODE"`
	DBLocalPath string `mapstructure:"DB_LOCAL_PATH"`
	NomeEmpresa string `mapstructure:"NOME_EMPRESA"`
}

func GetConfig() (Config, error) {
	var cfg Config

	viper.SetConfigName("signal-admin.conf")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return cfg, err
		}
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("parsing config data: %w", err)
	}

	if cfg.NomeEmpresa == "" {
		cfg.NomeEmpresa = "Signal Admin"
	}
	if cfg.DBMode == "" {
		cfg.DBMode = "sync"
	}
	if cfg.DBLocalPath == "" {
		cfg.DBLocalPath = "local.db"
	}

	return cfg, nil
}
