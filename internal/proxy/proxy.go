package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/Roman-Samoilenko/http-load-balancer/internal/balancer"
	"github.com/Roman-Samoilenko/http-load-balancer/internal/ratelimit"
	"github.com/Roman-Samoilenko/http-load-balancer/pkg/logger"
)

// LoadBalancer представляет основной сервис балансировщика нагрузки
type LoadBalancer struct {
	balancer     balancer.Balancer
	rateManager  *ratelimit.Manager
	reverseProxy *httputil.ReverseProxy
	logger       *logger.Logger
	server       *http.Server
}

// NewLoadBalancer создает новый экземпляр балансировщика нагрузки
func NewLoadBalancer(bal balancer.Balancer, rm *ratelimit.Manager, log *logger.Logger) *LoadBalancer {
	lb := &LoadBalancer{
		balancer:    bal,
		rateManager: rm,
		logger:      log,
	}

	// Создание обработчика запросов для ReverseProxy
	director := func(req *http.Request) {
		// Выбор бэкенда происходит в ServeHTTP
	}

	lb.reverseProxy = &httputil.ReverseProxy{
		Director: director,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			lb.logger.Error("Proxy error:", err)
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("Ошибка прокси: " + err.Error()))
		},
	}

	return lb
}

// ServeHTTP обрабатывает входящие HTTP-запросы
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Извлечение IP клиента для rate limiting
	clientIP := r.RemoteAddr

	// Проверка rate limit, если включен
	if lb.rateManager != nil {
		if !lb.rateManager.Allow(clientIP) {
			lb.logger.Warn("Rate limit превышен для", clientIP)
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("Слишком много запросов. Пожалуйста, попробуйте позже."))
			return
		}
	}

	// Выбор бэкенда через балансировщик
	backend := lb.balancer.NextBackend()
	if backend == nil {
		lb.logger.Error("Нет доступных бэкендов")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("Нет доступных бэкендов"))
		return
	}

	// Логирование запроса
	lb.logger.Info(fmt.Sprintf("Запрос от %s к %s перенаправлен на %s",
		r.RemoteAddr, r.URL.Path, backend.URL))

	// Подготовка URL для проксирования
	backendURL, err := url.Parse(backend.URL)
	if err != nil {
		lb.logger.Error("Некорректный URL бэкенда:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	proxyReq := r.Clone(r.Context())
	proxyReq.URL.Scheme = backendURL.Scheme
	proxyReq.URL.Host = backendURL.Host
	proxyReq.URL.Path = r.URL.Path
	proxyReq.URL.RawQuery = r.URL.RawQuery
	proxyReq.RequestURI = ""

	// Добавление заголовков прокси
	proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)
	proxyReq.Header.Set("X-Forwarded-Host", r.Host)
	proxyReq.Header.Set("X-Forwarded-Proto", r.URL.Scheme)

	// Проксирование запроса
	lb.reverseProxy.ServeHTTP(w, proxyReq)

	// Уменьшаем счетчик активных соединений
	backend.DecrementConnections()
}

// Start запускает HTTP-сервер
func (lb *LoadBalancer) Start(addr string) *http.Server {
	lb.server = &http.Server{
		Addr:    addr,
		Handler: lb,
	}

	// Запуск сервера в отдельной горутине
	go func() {
		if err := lb.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			lb.logger.Error("Ошибка запуска сервера:", err)
		}
	}()

	return lb.server
}

// Shutdown выполняет graceful shutdown сервера
func (lb *LoadBalancer) Shutdown(ctx context.Context) error {
	return lb.server.Shutdown(ctx)
}
