package librepayment

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	uuid "github.com/satori/go.uuid"
)

const notificationURLFieldName = "notification_url"

var ErrWrongPaymentStatus = errors.New("wrong payment status")

type LibrePayment struct {
	stor        *MemoryStorage
	notificator *Notificator
}

type Payment struct {
	CreatedAt time.Time
	ID        string
	Amount    float64
	Merchant  string

	// Pyaload is request fields except amount and merchant
	Payload map[string]string

	StatusHistory []StatusHistoryRecord

	realStatus PaymentStatus
}

type StatusHistoryRecord struct {
	EventAt time.Time
	Status  PaymentStatus
}

type PaymentStatus int

// New -> FormShowed -> Authorizing -> Authorized -> Confirming -> Confirmed
// | (New, FormShowed) -> DeadlineExpired
// | (New, FormShowed) -> Canceled
// | (Authorizing, Authorized, Confirming) -> Rejected
// | Confirmed -> Refunded
const (
	StatusNew PaymentStatus = iota + 1
	StatusFormShowed
	StatusDeadlineExpired
	StatusCanceled
	StatusAuthorizing
	StatusAuthorized
	StatusConfirming
	StatusConfirmed
	StatusRejected
	StatusRefunding
	StatusRefunded
)

type JournalRecord struct {
	TriedAt  time.Time
	TryNum   int
	Code     int
	Response []byte
	Error    string
}

func NewDefaultLibrePyament() *LibrePayment {
	notificator := NewNotificator(newNotificationJournal(), &http.Client{
		Timeout: time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // trus self signed / invalid certificate
			},
		},
	})
	notificator.Go()

	return &LibrePayment{
		stor:        NewMemoryStorage(),
		notificator: notificator,
	}
}

func (p *LibrePayment) Register(merchant string, amount float64, payload map[string]string) (string, error) {
	internalPaymentID := generateID()

	err := p.registerPayment(internalPaymentID, amount, merchant, payload)
	if err != nil {
		return "", err
	}

	return internalPaymentID, nil
}

func (p *LibrePayment) Status(id string) (Payment, error) {
	storagePayment, err := p.stor.Get(id)
	if err != nil {
		return Payment{}, err
	}

	return storagePaymentToEntity(storagePayment), nil
}

func (p *LibrePayment) Confirm(id string) error {
	var spc StoragePayment

	var opError error

	err := p.stor.WithPayment(id, func(sp *StoragePayment) {
		payment := storagePaymentToEntity(*sp)

		confirmErr := payment.Confirm()
		if confirmErr != nil {
			opError = confirmErr
			return
		}

		sp.setStatus(payment.currentStatus(), time.Now())
		spc = sp.clone()
	})
	if err != nil {
		return fmt.Errorf("cannot confirm payment: %w", err)
	}

	if opError != nil {
		return fmt.Errorf("cannot confirm payment: %w", opError)
	}

	p.sendNotificationIfNeeded(spc)

	return nil
}

func (p *LibrePayment) Reject(id string) error {
	return p.Cancel(id)
}

func (p *LibrePayment) Cancel(id string) error {
	var spc StoragePayment

	var opError error

	err := p.stor.WithPayment(id, func(sp *StoragePayment) {
		payment := storagePaymentToEntity(*sp)

		canelError := payment.Cancel()
		if canelError != nil {
			opError = canelError
			return
		}

		sp.setStatus(payment.currentStatus(), time.Now())
		spc = sp.clone()
	})
	if err != nil {
		return fmt.Errorf("cannot get payment for cancel: %w", err)
	}

	if opError != nil {
		return fmt.Errorf("cannot cancel payment: %w", opError)
	}

	p.sendNotificationIfNeeded(spc)

	return nil
}

func (p *LibrePayment) ForceSetStatus(id string, status PaymentStatus) error {
	var spc StoragePayment

	err := p.stor.WithPayment(id, func(sp *StoragePayment) {
		sp.setStatus(status, time.Now())
		spc = sp.clone()
	})

	if err != nil {
		return fmt.Errorf("cannot confirm payment: %w", err)
	}

	p.sendNotificationIfNeeded(spc)

	return nil
}

func (p *LibrePayment) AllPaymentsDescOrder() ([]Payment, error) {
	storagePayments, err := p.stor.GetAllDesc()
	if err != nil {
		return nil, err
	}

	payments := make([]Payment, len(storagePayments))

	for i, p := range storagePayments {
		payments[i] = storagePaymentToEntity(p)
	}

	return payments, nil
}

func (p *LibrePayment) Journal(id string) ([]JournalRecord, error) {
	recs, err := p.getJournalRecordsForId(id)
	if err != nil {
		if err == errJournalNotJound {
			return nil, nil
		}
		return nil, err
	}

	journal := make([]JournalRecord, len(recs))
	for i, rec := range recs {
		journal[i] = JournalRecord{
			TriedAt:  rec.triedAt,
			TryNum:   rec.tryNum,
			Code:     rec.responseCode,
			Response: rec.response,
			Error:    rec.err,
		}
	}

	return journal, nil
}

func (p *LibrePayment) registerPayment(id string, amount float64, merchant string, payload map[string]string) error {
	err := p.stor.Add(time.Now(), id, amount, merchant, payload)

	return err
}

func (p *LibrePayment) sendNotificationIfNeeded(sp StoragePayment) {
	notificationURL, ok := sp.Payload[notificationURLFieldName]
	if !ok || notificationURL == "" {
		return
	}

	// copy
	moreFields := make(map[string]string, len(sp.Payload))
	for k, v := range sp.Payload {
		moreFields[k] = v
	}

	p.notificator.Notify(notificationURL, Notification{
		ID:               sp.ID,
		Merchant:         sp.Merchant,
		Status:           sp.Status.String(),
		ErrorCode:        "0",
		NotReqiredFileds: moreFields,
	})
}

func (p *LibrePayment) getJournalRecordsForId(id string) ([]journalRecord, error) {
	return p.notificator.getJournalRecords(id)
}

func (p *Payment) Confirm() error {
	if !p.canBeConfirmed() {
		return ErrWrongPaymentStatus

	}

	// TODO: implement move to next status by SFM
	p.realStatus = StatusConfirmed

	return nil
}

func (p *Payment) Cancel() error {
	if !p.canBeCanceled() {
		return ErrWrongPaymentStatus
	}

	// TODO: implement move to next status by SFM
	switch p.realStatus {
	case StatusNew, StatusFormShowed:
		p.realStatus = StatusCanceled

	case StatusAuthorizing, StatusAuthorized, StatusConfirming:
		p.realStatus = StatusRejected

	case StatusConfirmed:
		p.realStatus = StatusRefunded

	default:
		return ErrWrongPaymentStatus
	}

	return nil
}

func (p *Payment) Status() string {
	return p.realStatus.String()
}

func (p *Payment) canBeConfirmed() bool {
	return p.realStatus == StatusNew || p.realStatus == StatusFormShowed
}

func (p *Payment) canBeCanceled() bool {
	return p.realStatus == StatusNew ||
		p.realStatus == StatusFormShowed ||
		p.realStatus == StatusAuthorizing ||
		p.realStatus == StatusAuthorized ||
		p.realStatus == StatusConfirming ||
		p.realStatus == StatusConfirmed
}

func (p *Payment) currentStatus() PaymentStatus {
	return p.realStatus
}

func generateID() string {
	return uuid.NewV4().String()
}

func storagePaymentToEntity(p StoragePayment) Payment {
	payment := Payment{
		CreatedAt:     p.CreatedAt,
		ID:            p.ID,
		Amount:        p.Amount,
		Merchant:      p.Merchant,
		realStatus:    p.Status,
		Payload:       map[string]string{},
		StatusHistory: make([]StatusHistoryRecord, len(p.StatusHistory)),
	}

	copy(payment.StatusHistory, p.StatusHistory)

	for k, v := range p.Payload {
		payment.Payload[k] = v
	}

	return payment
}
