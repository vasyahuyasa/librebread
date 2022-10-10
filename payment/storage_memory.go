package payment

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

type ProviderType string

const (
	TinkoffProvider  ProviderType = "Tinkoff"
	SberbankProvider ProviderType = "Sberbank"
)

func (pt *ProviderType) UnmarshalJSON(b []byte) error {
	var paymentProvider string
	if err := json.Unmarshal(b, &paymentProvider); err != nil {
		return err
	}

	if paymentProvider != string(TinkoffProvider) && paymentProvider != string(SberbankProvider) {
		log.Printf("UnmarshalJSON: %v", ErrIncorrectPaymentProvider)
	}

	*pt = ProviderType(paymentProvider)

	return nil
}

type MemoryStorage struct {
	storageMu                  sync.Mutex
	providers                  map[int64]Provider
	paymentProcesses           map[int64]PaymentProcess
	listProcessIDsInOrder      []int64
	clients                    map[string]Client
	clientIDsIndexesByRebillID map[int64]string
}

type PaymentProcess struct {
	ProcessID                      int64
	ProviderID                     int64
	CreatedOn                      time.Time
	PaymentURL                     string
	SuccessURL                     string
	FailURL                        string
	NotificationURL                string
	NotificationResponseOkReceived bool
	Status                         string
	Amount                         int64
	Description                    string
	Recurrent                      bool
	ClientID                       string
	OrderID                        string
}

type Client struct {
	ClientID   string
	RebillID   int64
	CardNumber string
}

type Provider struct {
	ProviderID int64
	Type       ProviderType
	Login      string
	Password   string
	URL        string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		providers:                  make(map[int64]Provider),
		paymentProcesses:           make(map[int64]PaymentProcess),
		clients:                    make(map[string]Client),
		clientIDsIndexesByRebillID: make(map[int64]string),
	}
}

// Payment Processes
func (s *MemoryStorage) HasProcessByID(processID int64) bool {
	_, ok := s.paymentProcesses[processID]
	return ok
}

func (s *MemoryStorage) GetProcessByID(processID int64) (PaymentProcess, error) {
	s.storageMu.Lock()
	defer s.storageMu.Unlock()

	if !s.HasProcessByID(processID) {
		return PaymentProcess{}, ErrPaymentProcessNotFound
	}
	return s.paymentProcesses[processID], nil
}

func (s *MemoryStorage) AddProcess(p PaymentProcess) error {
	s.storageMu.Lock()
	defer s.storageMu.Unlock()

	if s.HasProcessByID(p.ProcessID) {
		return ErrPaymentProcessAlreadyExists
	}

	s.paymentProcesses[p.ProcessID] = p
	s.listProcessIDsInOrder = append(s.listProcessIDsInOrder, p.ProcessID)
	return nil
}

func (s *MemoryStorage) UpdateProcess(p PaymentProcess) error {
	s.storageMu.Lock()
	defer s.storageMu.Unlock()

	if !s.HasProcessByID(p.ProcessID) {
		return ErrPaymentProcessNotFound
	}

	s.paymentProcesses[p.ProcessID] = p
	return nil
}

func (s *MemoryStorage) LastProcesses(n int) []PaymentProcess {
	s.storageMu.Lock()
	defer s.storageMu.Unlock()

	processes := make([]PaymentProcess, 0, n)
	length := len(s.paymentProcesses) - 1

	for i := 0; length-i >= 0 && i < n; i++ {
		processes = append(processes, s.paymentProcesses[s.listProcessIDsInOrder[length-i]])
	}
	return processes
}

func (s *MemoryStorage) ProcessesLen() int {
	return len(s.paymentProcesses)
}

// Client
func (s *MemoryStorage) HasClientByID(clientID string) bool {
	_, ok := s.clients[clientID]
	return ok
}

func (s *MemoryStorage) HasRebillID(rebillID int64) bool {
	_, ok := s.clientIDsIndexesByRebillID[rebillID]
	return ok
}

func (s *MemoryStorage) GetClientByID(clientID string) (Client, error) {
	s.storageMu.Lock()
	defer s.storageMu.Unlock()

	if !s.HasClientByID(clientID) {
		return Client{}, ErrClientNotFound
	}
	return s.clients[clientID], nil
}

func (s *MemoryStorage) AddClient(c Client) error {
	s.storageMu.Lock()
	defer s.storageMu.Unlock()

	if s.HasClientByID(c.ClientID) {
		return ErrClientAlreadyExists
	}

	if s.HasRebillID(c.RebillID) {
		return ErrRebillIDAlreadyExists
	}

	s.clients[c.ClientID] = c
	s.clientIDsIndexesByRebillID[c.RebillID] = c.ClientID
	return nil
}

// Provider
func (s *MemoryStorage) HasProviderByID(providerID int64) bool {
	_, ok := s.providers[providerID]
	return ok
}

func (s *MemoryStorage) HasProviderByLogin(login string) bool {
	for id := range s.providers {
		if s.providers[id].Login == login {
			return true
		}
	}
	return false
}

func (s *MemoryStorage) GetProviderByID(providerID int64) (Provider, error) {
	s.storageMu.Lock()
	defer s.storageMu.Unlock()

	if !s.HasProviderByID(providerID) {
		return Provider{}, ErrPaymentProviderNotFound
	}
	return s.providers[providerID], nil
}

func (s *MemoryStorage) GetProviderByLogin(login string) (Provider, error) {
	s.storageMu.Lock()
	defer s.storageMu.Unlock()

	for id := range s.providers {
		if s.providers[id].Login == login {
			return s.providers[id], nil
		}
	}
	return Provider{}, ErrPaymentProviderNotFound
}

func (s *MemoryStorage) AddProvider(p Provider) error {
	s.storageMu.Lock()
	defer s.storageMu.Unlock()

	if s.HasProviderByID(p.ProviderID) {
		return ErrPaymentProviderIDAlreadyExists
	}

	if s.HasProviderByLogin(p.Login) {
		return ErrPaymentProviderLoginAlreadyExists
	}

	s.providers[p.ProviderID] = p
	return nil
}
