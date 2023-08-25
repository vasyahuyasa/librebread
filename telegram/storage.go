package telegram

import "sync"

type TelegramRequest struct {
	Token   string
	Method  string
	Payload map[string]string
}

type TelegramStorage struct {
	requests []TelegramRequest
	mu       *sync.RWMutex
}

func (s *TelegramStorage) Add(r TelegramRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.requests = append(s.requests, r)
}

func (s *TelegramStorage) AllByDateDesc() []TelegramRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]TelegramRequest, len(s.requests))

	for i := len(s.requests) - 1; i >= 0; i-- {
		list = append(list, s.requests[i])
	}

	return list
}
