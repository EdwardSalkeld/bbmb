package queue

import (
	"sync"
	"time"
)

type Manager struct {
	queues          map[string]*Queue
	mu              sync.RWMutex
	timeoutCallback func(int)
}

func NewManager() *Manager {
	return &Manager{
		queues: make(map[string]*Queue),
	}
}

func (m *Manager) SetTimeoutCallback(cb func(int)) {
	m.timeoutCallback = cb
}

func (m *Manager) EnsureQueue(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.queues[name]; !exists {
		m.queues[name] = NewQueue(name)
	}
}

func (m *Manager) GetQueue(name string) (*Queue, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	q, exists := m.queues[name]
	return q, exists
}

func (m *Manager) QueueCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.queues)
}

func (m *Manager) StartTimeoutScanner(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			m.mu.RLock()
			queues := make([]*Queue, 0, len(m.queues))
			for _, q := range m.queues {
				queues = append(queues, q)
			}
			m.mu.RUnlock()

			totalTimedOut := 0
			for _, q := range queues {
				count := q.RequeueTimedOut()
				totalTimedOut += count
			}

			if totalTimedOut > 0 && m.timeoutCallback != nil {
				m.timeoutCallback(totalTimedOut)
			}
		}
	}()
}

func (m *Manager) GetAllQueues() []*Queue {
	m.mu.RLock()
	defer m.mu.RUnlock()

	queues := make([]*Queue, 0, len(m.queues))
	for _, q := range m.queues {
		queues = append(queues, q)
	}
	return queues
}
