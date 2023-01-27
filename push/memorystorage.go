package push

import (
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var ErrMessageNotFound = errors.New("message not found in store")

type storeMessage Message

type MemoryStorage struct {
	nextID func() string

	mutex    sync.Mutex
	byID     map[string]*storeMessage
	messages []storeMessage
}

func NewMemoryStorage() *MemoryStorage {
	lastID := int64(0)

	return &MemoryStorage{
		nextID: func() string {
			id := atomic.AddInt64(&lastID, 1)
			return strconv.FormatInt(id, 10)
		},
		byID: map[string]*storeMessage{},
	}
}

func (store *MemoryStorage) AddMessage(pushService string, rawMsg []byte, tokens []string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	id := store.nextID()

	msg := storeMessage{
		ID:          id,
		Time:        time.Now(),
		PushService: pushService,
		RawMsg:      rawMsg,
		Tokens:      tokens,
	}

	// reverse message order with fewer allocations
	store.messages = append(store.messages, storeMessage{})
	copy(store.messages[1:], store.messages)
	store.messages[0] = msg

	store.byID[id] = &store.messages[0]

	return nil
}

func (store *MemoryStorage) AllMessages() ([]Message, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	msgs := make([]Message, 0, len(store.messages))

	for _, m := range store.messages {
		msgs = append(msgs, storeToMessage(m))
	}

	return msgs, nil
}

func (store *MemoryStorage) ByID(id string) (Message, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	msg, ok := store.byID[id]
	if !ok {
		return Message{}, ErrMessageNotFound
	}

	return storeToMessage(*msg), nil
}

func storeToMessage(msg storeMessage) Message {
	return Message(msg)
}
