package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerPort     int    `env:"SERVER_PORT" envDefault:"8080"`
	DatabaseURL    string `env:"DATABASE_URL,required"`
	RedisAddr      string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	RedisPassword  string `env:"REDIS_PASSWORD" envDefault:""`
	MinIOEndpoint  string `env:"MINIO_ENDPOINT" envDefault:"localhost:9000"`
	MinIOAccessKey string `env:"MINIO_ACCESS_KEY" envDefault:"minioadmin"`
	MinIOSecretKey string `env:"MINIO_SECRET_KEY" envDefault:"minioadmin"`
	MinIOBucket    string `env:"MINIO_BUCKET" envDefault:"receipts"`
	MinIOUseSSL    bool   `env:"MINIO_USE_SSL" envDefault:"false"`
	OCRBaseURL     string `env:"OCR_BASE_URL" envDefault:"http://localhost:8081"`
	OCRQueueKey    string `env:"OCR_QUEUE_KEY" envDefault:"ocr:jobs"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	if cfg.ServerPort < 1024 || cfg.ServerPort > 65535 {
		return Config{}, fmt.Errorf("invalid SERVER_PORT %d (expected 1024-65535)", cfg.ServerPort)
	}
	return cfg, nil
}
