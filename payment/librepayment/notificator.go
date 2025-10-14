package librepayment

import (
	"bytes"
	"container/list"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"
)

const tickerTick = time.Second * 10

var errJournalNotJound = errors.New("journal not found")

type Notification struct {
	ID               string
	Merchant         string
	Status           string
	ErrorCode        string
	NotReqiredFileds map[string]string
}

type pendingNotification struct {
	url      string
	n        Notification
	waitTill time.Time
	tryNum   int
}

type Notificator struct {
	maxTries             int
	delay                time.Duration
	mu                   *sync.Mutex
	pendingNotifications *list.List
	journal              *notificationJournal
	client               *http.Client

	nchan chan pendingNotification
}

type journalRecord struct {
	tryNum       int
	triedAt      time.Time
	responseCode int
	response     []byte
	err          string
}

type notificationJournal struct {
	mu          *sync.Mutex
	records     []journalRecord
	byPaymentId map[string][]journalRecord
}

func NewNotificator(journal *notificationJournal, client *http.Client) *Notificator {
	return &Notificator{
		maxTries:             60, // 1 hour every minute
		delay:                time.Minute,
		mu:                   &sync.Mutex{},
		pendingNotifications: list.New(),
		journal:              journal,
		client:               client,
		nchan:                make(chan pendingNotification, 100),
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
	n.mu.Lock()
	defer n.mu.Unlock()

	n.tryRegisterNotification(notificationURL, notification, time.Time{}, 1)
	n.flush()
}

func (n *Notificator) Go() {
	go n.runTicker()
	go n.runNotificationWorker()
}

func (n *Notificator) tryRegisterNotification(notificationURL string, notification Notification, waitTill time.Time, tryNum int) {
	if tryNum > n.maxTries {
		n.writeErrJournal(notification.ID, tryNum, time.Now(), "max tries reached, notification discarded")
		return
	}

	n.pendingNotifications.PushBack(pendingNotification{
		url:      notificationURL,
		n:        notification,
		waitTill: waitTill,
		tryNum:   tryNum,
	})
}

func (n *Notificator) flush() {
	ready := n.drainReadyNotifications()

	for _, notification := range ready {
		select {
		case n.nchan <- notification:
		default:
			n.tryRegisterNotification(notification.url, notification.n, time.Now().Add(n.delay), notification.tryNum+1)
			n.writeErrJournal(notification.n.ID, notification.tryNum, time.Now(), "send queue overflowed")
		}

	}
}

func (n *Notificator) runTicker() {
	tick := time.NewTicker(tickerTick)

	for range tick.C {
		n.mu.Lock()
		n.flush()
		n.mu.Unlock()
	}
}

func (n *Notificator) runNotificationWorker() {
	for notification := range n.nchan {
		triedAt := time.Now()
		ok, resp := n.sendNotification(notification.url, notification.n)
		n.writejournal(notification.n.ID, notification.tryNum, triedAt, resp.code, resp.body, resp.err)

		// requeue
		if !ok {
			n.mu.Lock()
			n.tryRegisterNotification(notification.url, notification.n, time.Now().Add(n.delay), notification.tryNum+1)
			n.mu.Unlock()
		}
	}
}

type response struct {
	code int
	body []byte
	err  string
}

func (n *Notificator) sendNotification(url string, notification Notification) (bool, response) {
	payload := map[string]any{
		"id":        notification.ID,
		"merchant":  notification.Merchant,
		"status":    notification.Status,
		"errorCode": notification.ErrorCode,
	}

	for k, v := range notification.NotReqiredFileds {
		payload[k] = v
	}

	jsonStrPayload, err := json.Marshal(payload)
	if err != nil {
		return false, response{
			code: -1,
			body: nil,
			err:  err.Error(),
		}
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStrPayload))
	if err != nil {
		return false, response{
			code: -1,
			body: nil,
			err:  err.Error(),
		}
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return false, response{
			code: -1,
			body: []byte(""),
			err:  err.Error(),
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, response{
				code: resp.StatusCode,
				body: body,
				err:  "cannot read response body: " + err.Error(),
			}
		}

		return false, response{
			code: resp.StatusCode,
			body: body,
		}
	}

	body, err := io.ReadAll(resp.Body)

	return true, response{
		code: resp.StatusCode,
		body: body,
		err:  err.Error(),
	}
}

func (n *Notificator) drainReadyNotifications() []pendingNotification {
	var ready []pendingNotification

	for e := n.pendingNotifications.Front(); e != nil; e = e.Next() {
		pn, ok := e.Value.(pendingNotification)
		if !ok {
			panic("wrong type in pending notification list. Must investigate")
		}

		if time.Now().After(pn.waitTill) {
			ready = append(ready, pn)
			n.pendingNotifications.Remove(e)
		}
	}

	return ready
}

func (n *Notificator) writejournal(paymentId string, tryNum int, triedAt time.Time, responseCode int, response []byte, err string) {
	n.journal.write(paymentId, journalRecord{
		tryNum:       tryNum,
		triedAt:      triedAt,
		responseCode: responseCode,
		response:     response,
		err:          err,
	})
}

func (n *Notificator) writeErrJournal(paymentId string, tryNum int, triedAt time.Time, err string) {
	n.journal.write(paymentId, journalRecord{
		tryNum:       tryNum,
		triedAt:      triedAt,
		responseCode: -1,
		response:     []byte(""),
		err:          err,
	})
}

func (n *Notificator) getJournalRecords(id string) ([]journalRecord, error) {
	return n.journal.getForId(id)
}

func (j *notificationJournal) write(paymentId string, r journalRecord) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.records = append(j.records, r)

	list := j.byPaymentId[paymentId]
	list = append(list, r)
	j.byPaymentId[paymentId] = list
}

func (j *notificationJournal) getForId(id string) ([]journalRecord, error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	list, ok := j.byPaymentId[id]
	if !ok {
		return nil, errJournalNotJound
	}

	records := make([]journalRecord, len(list))
	copy(records, list)

	return records, nil
}
