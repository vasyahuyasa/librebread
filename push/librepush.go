package push

import (
	"fmt"
	"time"
)

type Message struct {
	ID          string
	Time        time.Time
	PushService string
	RawMsg      []byte
	Tokens      []string
}

type Storage interface {
	AddMessage(pushService string, rawMsg []byte, tokens []string) error
	AllMessages() ([]Message, error)
	ByID(id string) (Message, error)
}

type SendResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id"`
	Error     error  `json:"error"`
}

type BatchResponse struct {
	SuccessCount int             `json:"success_count"`
	FailureCount int             `json:"failure_count"`
	Responses    []*SendResponse `json:"responses"`
}

type LibrePush struct {
	storage Storage
}

func NewLibrePush(storage Storage) *LibrePush {
	return &LibrePush{
		storage: storage,
	}
}

func (p *LibrePush) Send(pushService string, rawMsg []byte, tokens []string) (*BatchResponse, error) {
	err := p.storage.AddMessage(pushService, rawMsg, tokens)
	if err != nil {
		return nil, fmt.Errorf("can not save batch message to storage: %w", err)
	}

	var response BatchResponse

	allSuccess(tokens, &response)

	return &response, nil
}

func (p *LibrePush) SendDryRun(tokens []string) (*BatchResponse, error) {
	var response BatchResponse

	allSuccess(tokens, &response)

	return &response, nil
}

func allSuccess(tokens []string, response *BatchResponse) {
	response.SuccessCount = len(tokens)
	response.FailureCount = 0

	for _, token := range tokens {
		response.Responses = append(response.Responses, &SendResponse{
			Success:   true,
			MessageID: token,
			Error:     nil,
		})
	}
}
