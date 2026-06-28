package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerPort    int    `env:"SERVER_PORT" envDefault:"8080"`
	DatabaseURL   string `env:"DATABASE_URL,required"`
	RedisAddr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`
	S3Endpoint    string `env:"S3_ENDPOINT" envDefault:"localhost:9000"`
	S3AccessKey   string `env:"S3_ACCESS_KEY" envDefault:"minioadmin"`
	S3SecretKey   string `env:"S3_SECRET_KEY" envDefault:"minioadmin"`
	S3Bucket      string `env:"S3_BUCKET" envDefault:"receipts"`
	S3UseSSL      bool   `env:"S3_USE_SSL" envDefault:"false"`
	OCRBaseURL    string `env:"OCR_BASE_URL" envDefault:"http://localhost:8081"`
	OCRQueueKey   string `env:"OCR_QUEUE_KEY" envDefault:"ocr:jobs"`
	JWTSecret     string `env:"JWT_SECRET" envDefault:"ahorrapp-dev-secret-change-me"`
	LogLevel      string `env:"LOG_LEVEL" envDefault:"info"`
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
