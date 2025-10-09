package librepayment

import (
	"fmt"
	"sync"
	"time"
)

type PaymentStatus int

type StoragePayment struct {
	CreatedAt time.Time
	ID        string
	Amount    float64
	Merchant  string
	Status    PaymentStatus
	Payload   map[string]string
	Meta      map[string]string
}

type MemoryStorage struct {
	payments     []*StoragePayment
	paymentsByID map[string]*StoragePayment

	mu *sync.RWMutex
}

const (
	StatusNew PaymentStatus = iota + 1
	StatusConfirmed
	StatusRejected
)

var (
	ErrPaymentAlreadyRegistered = fmt.Errorf("payment with id already registered")
	ErrPaymentNotFound          = fmt.Errorf("payment not found")
)

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		payments:     nil,
		paymentsByID: map[string]*StoragePayment{},
		mu:           &sync.RWMutex{},
	}
}

func (stor *MemoryStorage) Add(createdAt time.Time, id string, amount float64, merchant string, payload map[string]string) error {
	stor.mu.Lock()
	defer stor.mu.Unlock()

	_, ok := stor.paymentsByID[id]
	if ok {
		return ErrPaymentAlreadyRegistered
	}

	payment := &StoragePayment{
		CreatedAt: createdAt,
		ID:        id,
		Amount:    amount,
		Merchant:  merchant,
		Status:    StatusNew,
		Payload:   payload,
	}

	stor.payments = append(stor.payments, payment)
	stor.paymentsByID[id] = payment

	return nil
}

func (stor *MemoryStorage) SetPaymentStatus(id string, status PaymentStatus) error {
	stor.mu.Lock()
	defer stor.mu.Unlock()

	payment, ok := stor.paymentsByID[id]
	if !ok {
		return ErrPaymentNotFound
	}

	payment.Status = status

	return nil
}

func (stor *MemoryStorage) Get(id string) (StoragePayment, error) {
	stor.mu.RLock()
	defer stor.mu.RUnlock()

	p, ok := stor.paymentsByID[id]
	if !ok {
		return StoragePayment{}, ErrPaymentNotFound
	}

	return p.clone(), nil
}

func (stor *MemoryStorage) GetAllDesc() ([]StoragePayment, error) {
	stor.mu.RLock()
	defer stor.mu.RUnlock()

	numPayments := len(stor.payments)

	list := make([]StoragePayment, numPayments)

	for i, p := range stor.payments {
		list[numPayments-1-i] = p.clone()
	}

	return list, nil
}

func (payment *StoragePayment) clone() StoragePayment {
	p := StoragePayment{
		CreatedAt: payment.CreatedAt,
		ID:        payment.ID,
		Amount:    payment.Amount,
		Merchant:  payment.Merchant,
		Status:    payment.Status,
		Payload:   map[string]string{},
	}

	for k, v := range payment.Payload {
		p.Payload[k] = v
	}

	return p
}

func (status PaymentStatus) String() string {
	switch status {
	case StatusNew:
		return "new"

	case StatusConfirmed:
		return "confirmed"

	case StatusRejected:
		return "rejected"

	default:
		return fmt.Sprintf("unknown: %d", int(status))
	}
}
