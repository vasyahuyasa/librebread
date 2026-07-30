package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/vasyahuyasa/librebread/payment/librepayment"
)

func TestLibrePaymentRoutesExposeAllMermaidStatusSetters(t *testing.T) {
	payments := librepayment.NewDefaultLibrePyament()
	handler := librepayment.NewLibrePaymentHandler(payments, librepayment.NewPaymentUrlGeneratorFromENV())
	router := chi.NewRouter()
	librePaymentRoutes(router, handler)

	statuses := []string{
		"new",
		"formShowed",
		"deadlineExpired",
		"canceled",
		"authorizing",
		"authorized",
		"confirming",
		"confirmed",
		"rejected",
		"refunding",
		"refunded",
	}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			id, err := payments.Register("merchant-1", 150.25, nil)
			if err != nil {
				t.Fatalf("Register() error = %v", err)
			}

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/libre/payment/"+id+"/status/"+status, nil)
			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("status code = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
			}

			payment, err := payments.Status(id)
			if err != nil {
				t.Fatalf("Status() error = %v", err)
			}

			if payment.Status() != status {
				t.Fatalf("payment status = %q, want %q", payment.Status(), status)
			}
		})
	}
}
