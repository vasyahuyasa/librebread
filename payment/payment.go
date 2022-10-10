package payment

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
)

const (
	minProcessID = 1000
	maxProcessID = 10000000
	minRebillID  = 1000
	maxRebillID  = 10000000

	mockPaymentProvidersFile = "./mock_payment_providers.json"
)

type Payment struct {
	stor *MemoryStorage
}

func NewPayment() *Payment {
	return &Payment{
		stor: NewMemoryStorage(),
	}
}

func AddMockProviders(p *Payment) {
	type Providers struct {
		Providers []Provider `json:"Providers"`
	}
	var providers Providers

	jsonFile, err := os.Open(mockPaymentProvidersFile)
	if err != nil {
		log.Printf("Mock payment providers properties: %v", err)
		return
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Printf("Mock payment providers properties: %v", err)
		return
	}

	json.Unmarshal(byteValue, &providers)

	for i := 0; i < len(providers.Providers); i++ {
		if err := p.AddProvider(providers.Providers[i]); err != nil {
			log.Printf("Add mock payment providers: %v", err)
			return
		}

		log.Printf("Added mock payment providers: %v", providers.Providers[i].Login)
	}
}

// Payment Processes
func GetProcessIDFromURL(r *http.Request) (processID int64, err error) {
	processID, err = strconv.ParseInt(chi.URLParam(r, "processID"), 10, 64)

	if err != nil {
		return processID, err
	}

	if processID == 0 {
		return processID, ErrRequestNotProcessID
	}

	return processID, nil
}

func (p *Payment) HasProcessByID(processID int64) bool {
	return p.stor.HasProcessByID(processID)
}

func (p *Payment) GetProcessByID(processID int64) (PaymentProcess, error) {
	return p.stor.GetProcessByID(processID)
}

func (p *Payment) AddProcess(paymentProcess PaymentProcess) error {
	return p.stor.AddProcess(paymentProcess)
}

func (p *Payment) UpdateProcess(paymentProcess PaymentProcess) error {
	return p.stor.UpdateProcess(paymentProcess)
}

func (p *Payment) LastProcesses(n int) []PaymentProcess {
	return p.stor.LastProcesses(n)
}

func (p *Payment) ProcessesLen() int {
	return p.stor.ProcessesLen()
}

func (p *Payment) ConvertToTemplatePaymentProcess(paymentProcess PaymentProcess) (tplPayment TemplatePaymentProcess, err error) {
	provider, err := p.GetProviderByID(paymentProcess.ProviderID)
	if err != nil {
		return TemplatePaymentProcess{}, err
	}

	tplPayment = TemplatePaymentProcess{
		ProcessID:                      paymentProcess.ProcessID,
		Provider:                       provider.Type,
		CreatedOn:                      paymentProcess.CreatedOn,
		PaymentURL:                     paymentProcess.PaymentURL,
		SuccessURL:                     paymentProcess.SuccessURL,
		FailURL:                        paymentProcess.FailURL,
		NotificationURL:                paymentProcess.NotificationURL,
		NotificationResponseOkReceived: paymentProcess.NotificationResponseOkReceived,
		Status:                         paymentProcess.Status,
		Amount:                         paymentProcess.Amount,
		Description:                    paymentProcess.Description,
		Recurrent:                      paymentProcess.Recurrent,
		ClientID:                       paymentProcess.ClientID,
		OrderID:                        paymentProcess.OrderID,
	}

	return tplPayment, nil
}

func (p *Payment) ConvertToTemplatePaymentProcesses(paymentProcesses []PaymentProcess) (tplPayments []TemplatePaymentProcess, err error) {
	for _, paymentProcess := range paymentProcesses {
		tplPayment, err := p.ConvertToTemplatePaymentProcess(paymentProcess)
		if err != nil {
			return []TemplatePaymentProcess{}, err
		}

		tplPayments = append(tplPayments, tplPayment)
	}

	return tplPayments, nil
}

// Client
func (p *Payment) HasClientByID(clientID string) bool {
	return p.stor.HasClientByID(clientID)
}

func (p *Payment) HasRebillID(rebillID int64) bool {
	return p.stor.HasRebillID(rebillID)
}

func (p *Payment) GetClientByID(clientID string) (Client, error) {
	return p.stor.GetClientByID(clientID)
}

func (p *Payment) CreateClient(clientID string) (Client, error) {
	client := Client{
		ClientID:   clientID,
		RebillID:   p.generateRebillID(),
		CardNumber: "430000******" + p.generate4Digits(),
	}

	return client, p.stor.AddClient(client)
}

func (p *Payment) generateRebillID() int64 {
	rebillID := rand.Int63n(maxRebillID-minRebillID) + minRebillID

	if p.stor.HasRebillID(rebillID) {
		return p.generateRebillID()
	}

	return rebillID
}

func (p *Payment) generate4Digits() (str4digits string) {
	for i := 0; i < 4; i++ {
		str4digits += strconv.Itoa(rand.Intn(10))
	}

	return str4digits
}

// Provider
func (p *Payment) HasProviderByID(providerID int64) bool {
	return p.stor.HasProviderByID(providerID)
}

func (p *Payment) HasProviderByLogin(login string) bool {
	return p.stor.HasProviderByLogin(login)
}

func (p *Payment) GetProviderByID(providerID int64) (Provider, error) {
	return p.stor.GetProviderByID(providerID)
}

func (p *Payment) GetProviderByLogin(login string) (Provider, error) {
	return p.stor.GetProviderByLogin(login)
}

func (p *Payment) AddProvider(provider Provider) error {
	return p.stor.AddProvider(provider)
}

func (p *Payment) GenerateProcessID() int64 {
	processID := rand.Int63n(maxProcessID-minProcessID) + minProcessID

	if p.stor.HasProcessByID(processID) {
		return p.GenerateProcessID()
	}

	return processID
}
