package librepayment

import (
	"fmt"
	"sync"
	"time"
)

type StoragePayment struct {
	CreatedAt     time.Time
	ID            string
	Amount        float64
	Merchant      string
	Status        PaymentStatus
	StatusHistory []StatusHistoryRecord
	Payload       map[string]string
	Meta          map[string]string
}

type MemoryStorage struct {
	payments     []*StoragePayment
	paymentsByID map[string]*StoragePayment

	mu *sync.RWMutex
}

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
		StatusHistory: []StatusHistoryRecord{
			{
				EventAt: createdAt,
				Status:  StatusNew,
			},
		},
		Payload: payload,
	}

	stor.payments = append(stor.payments, payment)
	stor.paymentsByID[id] = payment

	return nil
}

func (stor *MemoryStorage) Get(id string) (StoragePayment, error) {
	stor.mu.RLock()
	defer stor.mu.RUnlock()

	p, err := stor.get(id)
	if err != nil {
		return StoragePayment{}, err
	}

	return p.clone(), nil
}

// WithPayment retrieves a payment by ID and runs the given function with it.
// It uses a write lock because callers mutate the stored payment.
func (stor *MemoryStorage) WithPayment(id string, f func(*StoragePayment)) error {
	stor.mu.Lock()
	defer stor.mu.Unlock()

	p, err := stor.get(id)
	if err != nil {
		return err
	}

	f(p)

	return nil
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

func (stor *MemoryStorage) get(id string) (*StoragePayment, error) {
	p, ok := stor.paymentsByID[id]
	if !ok {
		return nil, ErrPaymentNotFound
	}

	return p, nil
}

func (payment *StoragePayment) setStatus(status PaymentStatus, eventAt time.Time) {
	if payment.Status == status {
		return
	}

	payment.Status = status
	payment.StatusHistory = append(payment.StatusHistory, StatusHistoryRecord{
		EventAt: eventAt,
		Status:  status,
	})
}

func (payment *StoragePayment) clone() StoragePayment {
	p := StoragePayment{
		CreatedAt:     payment.CreatedAt,
		ID:            payment.ID,
		Amount:        payment.Amount,
		Merchant:      payment.Merchant,
		Status:        payment.Status,
		StatusHistory: make([]StatusHistoryRecord, len(payment.StatusHistory)),
		Payload:       map[string]string{},
	}

	copy(p.StatusHistory, payment.StatusHistory)

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

	case StatusFormShowed:
		return "formShowed"

	case StatusDeadlineExpired:
		return "deadlineExpired"

	case StatusCanceled:
		return "canceled"

	case StatusAuthorizing:
		return "authorizing"

	case StatusAuthorized:
		return "authorized"

	case StatusConfirming:
		return "confirming"

	case StatusRefunding:
		return "refunding"

	case StatusRefunded:
		return "refunded"

	default:
		return fmt.Sprintf("unknown: %d", int(status))
	}
}
