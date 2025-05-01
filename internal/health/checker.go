package health

import (
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Roman-Samoilenko/http-load-balancer/internal/balancer"
	"github.com/Roman-Samoilenko/http-load-balancer/pkg/logger"
)

// Checker отвечает за проверку здоровья бэкендов
type Checker struct {
	balancer      balancer.Balancer
	checkInterval time.Duration
	timeout       time.Duration
	logger        *logger.Logger
	stopCh        chan struct{}
	wg            sync.WaitGroup
	httpClient    *http.Client
}

// NewChecker создает новый экземпляр Checker
func NewChecker(balancer balancer.Balancer, checkInterval, timeout time.Duration, logger *logger.Logger) *Checker {
	httpClient := &http.Client{
		Timeout: timeout * time.Second,
	}

	return &Checker{
		balancer:      balancer,
		checkInterval: checkInterval * time.Second,
		timeout:       timeout * time.Second,
		logger:        logger,
		stopCh:        make(chan struct{}),
		httpClient:    httpClient,
	}
}

// Start запускает периодические проверки бэкендов
func (c *Checker) Start() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		// Запускаем первую проверку сразу
		c.checkAllBackends()

		ticker := time.NewTicker(c.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.checkAllBackends()
			case <-c.stopCh:
				return
			}
		}
	}()
}

// Stop останавливает проверки
func (c *Checker) Stop() {
	close(c.stopCh)
	c.wg.Wait()
}

// checkAllBackends проверяет все бэкенды
func (c *Checker) checkAllBackends() {
	backends := c.balancer.Backends()

	// Проверяем каждый бэкенд в отдельной горутине
	var wg sync.WaitGroup
	for _, backend := range backends {
		wg.Add(1)
		go func(b *balancer.Backend) {
			defer wg.Done()
			currentStatus := b.IsAlive
			newStatus := c.checkBackend(b)

			// Если статус изменился, логируем
			if currentStatus != newStatus {
				if newStatus {
					c.logger.Info("Бэкенд восстановлен:", b.URL)
					c.balancer.MarkBackendUp(b.URL)
				} else {
					c.logger.Warn("Бэкенд недоступен:", b.URL)
					c.balancer.MarkBackendDown(b.URL)
				}
			}
		}(backend)
	}

	wg.Wait()
}

// checkBackend проверяет доступность одного бэкенда
func (c *Checker) checkBackend(backend *balancer.Backend) bool {
	backendURL, err := url.Parse(backend.URL)
	if err != nil {
		c.logger.Error("Некорректный URL бэкенда:", err)
		return false
	}

	// Формируем URL для проверки
	healthURL := *backendURL
	healthURL.Path = "/health" // Можно настроить путь для проверки

	// Отправляем GET запрос для проверки доступности
	resp, err := c.httpClient.Get(healthURL.String())

	if err != nil {
		// Проверяем, что ошибка имеет тип *url.Error
		if urlErr, ok := err.(*url.Error); ok {
			// Проверяем, что это ошибка таймаута
			if urlErr.Timeout() {
				// Обработка ошибки таймаута
				c.logger.Error("Произошёл таймаут запроса")
			} else {
				// Другая ошибка
				c.logger.Error("Ошибка запроса:", err)
			}
		} else {
			// Ошибка другого типа
			c.logger.Error("Ошибка запроса:", err)
		}
		return false
	}

	defer resp.Body.Close()

	// Считаем бэкенд здоровым, если код ответа 2xx
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
