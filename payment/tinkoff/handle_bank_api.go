package tinkoff

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/vasyahuyasa/librebread/payment"
)

func InitHandler(p *payment.Payment) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := initRequest(r)
		if err != nil {
			log.Printf("Tinkoff: %v", err)
			return
		}

		provider, err := p.GetProviderByLogin(input.TerminalKey)
		if err != nil {
			log.Printf("Tinkoff: %v", err)
			return
		}

		if provider.Type != payment.TinkoffProvider {
			log.Printf("Tinkoff: %v", payment.ErrIncorrectPaymentProvider)
			return
		}

		var client payment.Client
		if input.CustomerKey != "" {
			if p.HasClientByID(string(input.CustomerKey)) {
				client, err = p.GetClientByID(string(input.CustomerKey))
				if err != nil {
					log.Printf("Tinkoff: %v", err)
					return
				}
			} else {
				client, err = p.CreateClient(string(input.CustomerKey))
				if err != nil {
					log.Printf("Tinkoff: %v", err)
					return
				}
			}
		}

		isRecurrent := false
		if input.Recurrent == "Y" {
			isRecurrent = true
		}

		processID := p.GenerateProcessID()

		process := payment.PaymentProcess{
			ProcessID:       processID,
			ProviderID:      provider.ProviderID,
			CreatedOn:       time.Now(),
			PaymentURL:      fmt.Sprintf("%s/payment/%d", provider.URL, processID),
			SuccessURL:      input.SuccessURL,
			FailURL:         input.FailURL,
			NotificationURL: input.NotificationURL,
			Status:          string(StatusNew),
			Amount:          input.Amount,
			Description:     input.Description,
			Recurrent:       isRecurrent,
			ClientID:        client.ClientID,
			OrderID:         string(input.OrderID),
		}

		p.AddProcess(process)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(initResponse(process, input.TerminalKey))
	}
}

func ChargeHandler(p *payment.Payment) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := chargeRequest(r)
		if err != nil {
			log.Printf("Tinkoff: %v", err)
			return
		}

		process, provider, err := getPaymentProcessAndProviderByProcessID(input.ProcessID, p)
		if err != nil {
			log.Printf("Tinkoff: %v", err)
			return
		}

		process.NotificationResponseOkReceived = false
		process.Status = string(StatusConfirmed)

		client, _ := p.GetClientByID(process.ClientID)
		ok, err := sendNotification(process, client, provider)
		if err != nil {
			log.Printf("Tinkoff: %v", err)
		} else if ok {
			process.NotificationResponseOkReceived = true
		} else {
			log.Println("Tinkoff: incorrect response to the notification was received")
		}

		p.UpdateProcess(process)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chargeResponse(process, input.TerminalKey))
	}
}

func GetStateHandler(p *payment.Payment) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := getStateRequest(r)
		if err != nil {
			log.Printf("Tinkoff: %v", err)
			return
		}

		process, _, err := getPaymentProcessAndProviderByProcessID(input.ProcessID, p)
		if err != nil {
			if err == payment.ErrPaymentProcessNotFound {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(getStateResponsePaymentNotFound())
			} else {
				log.Printf("Tinkoff: %v", err)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(getStateResponseSuccess(process, input.TerminalKey))
	}
}

func getPaymentProcessAndProviderByProcessID(processID int64, p *payment.Payment) (process payment.PaymentProcess, provider payment.Provider, err error) {
	if processID == 0 {
		return process, provider, payment.ErrRequestNotProcessID
	}

	process, err = p.GetProcessByID(processID)
	if err != nil {
		return process, provider, err
	}

	provider, err = p.GetProviderByID(process.ProviderID)
	if err != nil {
		return process, provider, err
	}

	if provider.Type != payment.TinkoffProvider {
		return process, provider, err
	}

	return process, provider, nil
}
