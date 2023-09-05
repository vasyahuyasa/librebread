package telegram

import (
	"sync"
	"time"
)

type TelegramRequest struct {
	Time    time.Time
	Token   string
	Method  string
	Payload map[string]string
}

type MemoryTelegramStorage struct {
	requests []TelegramRequest
	mu       *sync.RWMutex
}

func NewMemoryTelegramStorage() *MemoryTelegramStorage {
	return &MemoryTelegramStorage{
		mu: &sync.RWMutex{},
	}
}

func (s *MemoryTelegramStorage) Add(r TelegramRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.requests = append(s.requests, r)

	return nil
}

func (s *MemoryTelegramStorage) AllSortedByDateDesc() ([]TelegramRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]TelegramRequest, 0, len(s.requests))

	for i := len(s.requests) - 1; i >= 0; i-- {
		list = append(list, s.requests[i])
	}

	return list, nil
}
