package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	. "github.com/Roman-Samoilenko/http-load-balancer/internal/backend"
	"github.com/Roman-Samoilenko/http-load-balancer/internal/balancer"
	"github.com/Roman-Samoilenko/http-load-balancer/internal/config"
	"github.com/Roman-Samoilenko/http-load-balancer/internal/health"
	"github.com/Roman-Samoilenko/http-load-balancer/internal/proxy"
	"github.com/Roman-Samoilenko/http-load-balancer/internal/ratelimit"
	"github.com/Roman-Samoilenko/http-load-balancer/pkg/logger"
)

func main() {
	// Разбор аргументов командной строки
	configPath := flag.String("config", "configs/config.json", "путь к конфигурационному файлу")
	flag.Parse()

	// Инициализация логгера
	log := logger.New("info")
	log.Info("Запуск балансировщика нагрузки")

	// Загрузка конфигурации
	cfg, err := config.LoadConfig(*configPath)

	if err != nil {
		log.Error("Ошибка загрузки конфигурации:", err)
		os.Exit(1)
	}
	log.Info("Конфигурация загружена успешно")

	// Создание бэкендов
	var backends []*Backend
	for _, backendCfg := range cfg.Backends {
		backend := &Backend{
			URL:     backendCfg.URL,
			Weight:  backendCfg.Weight,
			IsAlive: true,
		}
		backends = append(backends, backend)
		log.Info("Добавлен бэкенд: ", backendCfg.URL)
	}

	// Создание балансировщика
	var bal balancer.Balancer
	switch cfg.BalancerType {
	case "round-robin":
		bal = balancer.NewRoundRobin(backends)
		log.Info("Используется алгоритм балансировки Round-Robin")
	case "random":
		bal = balancer.NewRandom(backends)
		log.Info("Используется алгоритм балансировки Random")
	default:
		log.Info("Используется алгоритм балансировки Round-Robin (по умолчанию)")
		bal = balancer.NewRoundRobin(backends)
	}

	// Настройка rate limiting
	var rateLimiter *ratelimit.Manager
	if cfg.RateLimit.Enabled {
		rateLimiter = ratelimit.NewManager(cfg.RateLimit.DefaultCapacity, cfg.RateLimit.DefaultRate)
		log.Info("Rate limiting включен. Стандартный лимит: ", cfg.RateLimit.DefaultRate, " запросов в секунду")
	}

	// Настройка health checker
	healthChecker := health.NewChecker(bal, cfg.HealthCheck.Interval, cfg.HealthCheck.Timeout, log)
	healthChecker.Start()
	log.Info("Запущена проверка доступности бэкендов")

	// Создание прокси
	prx := proxy.NewLoadBalancer(bal, rateLimiter, log)

	// Запуск сервера
	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	server := prx.Start(serverAddr)
	log.Info("Сервер запущен на ", serverAddr)

	// Обработка сигналов для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Получен сигнал остановки, выполняется graceful shutdown...")

	// Остановка health checker
	healthChecker.Stop()

	// Graceful shutdown сервера
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Ошибка при остановке сервера:", err)
	}

	log.Info("Сервер остановлен")
}
