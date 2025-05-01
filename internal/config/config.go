package config

import (
	"encoding/json"
	"os"
	"time"
)

// Config представляет основную конфигурацию приложения
type Config struct {
	Server       ServerConfig      `json:"server"`
	Backends     []BackendConfig   `json:"backends"`
	BalancerType string            `json:"balancer_type"`
	RateLimit    RateLimitConfig   `json:"rate_limit"`
	HealthCheck  HealthCheckConfig `json:"health_check"`
}

// ServerConfig содержит настройки HTTP-сервера
type ServerConfig struct {
	Port int `json:"port"`
}

// BackendConfig содержит настройки бэкенд-сервера
type BackendConfig struct {
	URL      string `json:"url"`
	Weight   int    `json:"weight"`
	MaxConns int    `json:"max_connections"`
}

// RateLimitConfig содержит настройки ограничения частоты запросов
type RateLimitConfig struct {
	Enabled         bool    `json:"enabled"`
	DefaultRate     float64 `json:"default_rate"`
	DefaultCapacity int     `json:"default_capacity"`
}

// HealthCheckConfig содержит настройки проверки доступности бэкендов
type HealthCheckConfig struct {
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
}

// LoadConfig загружает конфигурацию из JSON-файла
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	// Установка значений по умолчанию, если они не указаны
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.BalancerType == "" {
		config.BalancerType = "round-robin"
	}
	if config.RateLimit.DefaultRate == 0 {
		config.RateLimit.DefaultRate = 10
	}
	if config.RateLimit.DefaultCapacity == 0 {
		config.RateLimit.DefaultCapacity = 100
	}
	if config.HealthCheck.Interval == 0 {
		config.HealthCheck.Interval = 10 * time.Second
	}
	if config.HealthCheck.Timeout == 0 {
		config.HealthCheck.Timeout = 2 * time.Second
	}

	return &config, nil
}
