package payment

import (
	"errors"
)

var (
	ErrPaymentProcessNotFound      = errors.New("Payment process not found")
	ErrPaymentProcessAlreadyExists = errors.New("Payment process already exists")
	ErrUnknownPaymentProcessStatus = errors.New("Unknown payment process status")

	ErrPaymentProviderNotFound           = errors.New("Payment provider not found")
	ErrPaymentProviderIDAlreadyExists    = errors.New("Payment provider with this ID already exists")
	ErrPaymentProviderLoginAlreadyExists = errors.New("Payment provider with this login already exists")
	ErrIncorrectPaymentProvider          = errors.New("Incorrect payment provider")

	ErrClientNotFound        = errors.New("Client not found")
	ErrClientAlreadyExists   = errors.New("Client already exists")
	ErrRebillIDAlreadyExists = errors.New("RebillID already exists")

	ErrRequestNotProcessID = errors.New("ProcessID param is required and must be a number")
)
