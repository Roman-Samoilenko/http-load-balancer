# Load Balancer

Простой балансировщик нагрузки на Go, который распределяет HTTP-запросы между несколькими бэкенд-серверами с использованием алгоритма Round-Robin и включает ограничение частоты запросов (rate limiting) с помощью алгоритма Token Bucket.

## Возможности

- Балансировка нагрузки с использованием алгоритма Round-Robin, least connections, random
- Ограничение частоты запросов (rate limiting) с использованием алгоритма Token Bucket
- Проверка доступности бэкенд-серверов (health check)
- Конфигурация через JSON-файл
- Логирование событий


## Установка

```bash
git clone https://github.com/Roman-Samoilenko/http-load-balancer.git
cd load-balancer
go build -o load-balancer ./cmd/server
```

## Конфигурация

Конфигурация приложения осуществляется через файл `configs/config.json`:

```json
{
  "server": {
    "port": 8080
  },
  "balancer_type": "round-robin",
  "backends": [
    {
      "url": "http://localhost:8081",
      "weight": 1,
      "max_connections": 100
    },
    {
      "url": "http://localhost:8082",
      "weight": 1,
      "max_connections": 100
    }
  ],
  "rate_limit": {
    "enabled": true,
    "default_rate": 10,
    "default_capacity": 100
  },
  "health_check": {
    "interval": 10,
    "timeout": 2
  }
}
```

### Параметры конфигурации

- `server.port` - порт, на котором будет работать балансировщик
- `balancer_type` - алгоритм балансировки (Round-Robin, least connections, random)
- `weight` - вес бэкенда для алгоритмов балансировки нагрузки

#### Rate limit:
- `default_rate` - скорость пополнения токенов для пользователя
- `default_capacity` - максимальный запас токенов для пользователя

#### Health Check:
- `interval` - временные промежутки проверки доступности бэкенда
- `timeout` - предельное время ожидания ответа
