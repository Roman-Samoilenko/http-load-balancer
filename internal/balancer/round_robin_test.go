package balancer

import (
	"testing"

	. "github.com/Roman-Samoilenko/http-load-balancer/internal/backend"
)

func TestRoundRobinBalancer(t *testing.T) {
	// Создаем два бэкенда для теста
	backend1 := &Backend{URL: "http://server1:8080", IsAlive: true}
	backend2 := &Backend{URL: "http://server2:8080", IsAlive: true}

	// Инициализируем балансировщик с двумя бэкендами
	balancer := NewRoundRobin([]*Backend{backend1, backend2})

	// Тест 1: Проверяем, что балансировщик возвращает бэкенды поочередно
	firstSelected := balancer.NextBackend()
	if firstSelected == nil {
		t.Fatalf("Ожидался бэкенд, получен nil")
	}

	secondSelected := balancer.NextBackend()
	if secondSelected == nil {
		t.Fatalf("Ожидался бэкенд, получен nil")
	}

	// Проверяем, что выбраны разные бэкенды
	if firstSelected == secondSelected {
		t.Errorf("Round-Robin не работает: получен один и тот же бэкенд дважды подряд")
	}

	// Тест 2: Проверяем, что счетчики соединений увеличиваются
	if firstSelected.GetActiveConnections() != 1 {
		t.Errorf("Счетчик соединений не инкрементирован для первого бэкенда")
	}

	if secondSelected.GetActiveConnections() != 1 {
		t.Errorf("Счетчик соединений не инкрементирован для второго бэкенда")
	}

	// Уменьшаем счетчики соединений
	firstSelected.DecrementConnections()
	secondSelected.DecrementConnections()

	// Проверяем, что счетчики обнулились
	if firstSelected.GetActiveConnections() != 0 {
		t.Errorf("Счетчик соединений не декрементирован для первого бэкенда")
	}

	if secondSelected.GetActiveConnections() != 0 {
		t.Errorf("Счетчик соединений не декрементирован для второго бэкенда")
	}

	// Тест 3: Проверка статуса доступности (alive)
	balancer.MarkBackendDown(backend1.URL)

	// После маркировки первого бэкенда как недоступного, должен всегда возвращаться второй
	for i := 0; i < 3; i++ {
		selected := balancer.NextBackend()
		if selected != backend2 {
			t.Errorf("Выбран недоступный бэкенд или некорректный бэкенд")
		}
	}

	// Проверяем, что бэкенд помечен как недоступный
	backend1.Mu.RLock()
	if backend1.IsAlive {
		t.Errorf("Бэкенд все еще помечен как доступный после MarkBackendDown")
	}
	backend1.Mu.RUnlock()

	// Тест 4: Возвращаем первый бэкенд обратно
	balancer.MarkBackendUp(backend1.URL)

	// Проверяем, что бэкенд помечен как доступный
	backend1.Mu.RLock()
	if !backend1.IsAlive {
		t.Errorf("Бэкенд не помечен как доступный после MarkBackendUp")
	}
	backend1.Mu.RUnlock()

	// После возвращения бэкенда должно продолжаться чередование
	firstAfterUp := balancer.NextBackend()
	secondAfterUp := balancer.NextBackend()

	if firstAfterUp == secondAfterUp {
		t.Errorf("Round-Robin не работает после восстановления бэкенда")
	}

	// Тест 5: Проверка добавления бэкенда
	backend3 := &Backend{URL: "http://server3:8080"}
	balancer.AddBackend(backend3)

	// Проверяем, что количество бэкендов увеличилось
	backends := balancer.Backends()
	if len(backends) != 3 {
		t.Errorf("Ожидалось 3 бэкенда после добавления, получено %d", len(backends))
	}

	// Тест 6: Проверка удаления бэкенда
	balancer.RemoveBackend(backend3.URL)

	// Проверяем, что количество бэкендов уменьшилось
	backends = balancer.Backends()
	if len(backends) != 2 {
		t.Errorf("Ожидалось 2 бэкенда после удаления, получено %d", len(backends))
	}

	// Тест 7: Проверка случая, когда все бэкенды недоступны
	balancer.MarkBackendDown(backend1.URL)
	balancer.MarkBackendDown(backend2.URL)

	// Проверяем, что возвращается nil, когда нет доступных бэкендов
	noBackend := balancer.NextBackend()
	if noBackend != nil {
		t.Errorf("Ожидался nil при отсутствии доступных бэкендов")
	}
}
