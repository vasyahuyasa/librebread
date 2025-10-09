package librepayment

import (
	"sync"
	"time"
)

type Notification struct {
	Merchant  string
	Success   bool
	Status    string
	PaymentId string
	ErrorCode string
}

type Notificator struct {
	maxTries int
	delay    time.Duration
	journal  *notificationJournal
}

type journalRecord struct {
	notification Notification
	tryNum       int
	triedAt      time.Time
	ok           bool
	responseCode int
	response     []byte
}

type notificationJournal struct {
	mu          *sync.Mutex
	records     []journalRecord
	byPaymentId map[string][]journalRecord
}

func NewNotificator(journal *notificationJournal) *Notificator {
	return &Notificator{
		maxTries: 5,
		delay:    time.Minute,
		journal:  journal,
	}
}

func newNotificationJournal() *notificationJournal {
	return &notificationJournal{
		mu:          &sync.Mutex{},
		records:     []journalRecord{},
		byPaymentId: map[string][]journalRecord{},
	}
}

func (n *Notificator) Notify(notificationURL string, notification Notification) {
	// TODO add to pending notification list, and call signal for send pending notifications
	n.registerNotification(notificationURL, notification)
}

func (n *Notificator) registerNotification(notificationURL string, notification Notification) {
	// TODO  store in pending list
}

func (j *notificationJournal) write(paymentId string, r journalRecord) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.records = append(j.records, r)

	list := j.byPaymentId[paymentId]
	list = append(list, r)
	j.byPaymentId[paymentId] = list
}
