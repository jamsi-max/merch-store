package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	JWTSecret   string `mapstructure:"JWT_SECRET"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("DATABASE_URL", "postgres://admin:secrets@db:5432/merch_store?sslmode=disable")
	viper.SetDefault("JWT_SECRET", "super_puper_mega_secrets_key_jwt")

	if err := viper.ReadInConfig(); err != nil {
		log.Println("[ERR] no .env file found, using default values or environment variables")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
