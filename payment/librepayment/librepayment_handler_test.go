package librepayment

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestGetPaymentStatusDoesNotReturnStatusHistory(t *testing.T) {
	payments := newTestLibrePayment()

	id, err := payments.Register("merchant-1", 150.25, nil)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	err = payments.Confirm(id)
	if err != nil {
		t.Fatalf("Confirm() error = %v", err)
	}

	handler := NewLibrePaymentHandler(payments, &PaymentUrlGenerator{baseURL: "http://localhost"})
	router := chi.NewRouter()
	router.Get("/libre/payment/{payment_id}", handler.GetPaymentStatus)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/libre/payment/"+id, nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var response map[string]any

	err = json.NewDecoder(recorder.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if response["id"] != id {
		t.Fatalf("response id = %q, want %q", response["id"], id)
	}

	if response["status"] != "confirmed" {
		t.Fatalf("response status = %q, want confirmed", response["status"])
	}

	if _, ok := response["status_history"]; ok {
		t.Fatalf("status_history must not be returned by status API: %+v", response["status_history"])
	}
}

func TestPaymentPageShowsStatusHistory(t *testing.T) {
	payments := newTestLibrePayment()

	id, err := payments.Register("merchant-1", 150.25, nil)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	err = payments.Confirm(id)
	if err != nil {
		t.Fatalf("Confirm() error = %v", err)
	}

	handler := NewLibrePaymentHandler(payments, &PaymentUrlGenerator{baseURL: "http://localhost"})
	router := chi.NewRouter()
	router.Get("/librepayments/{payment_id}", handler.PaymentPage)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/librepayments/"+id, nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	body := recorder.Body.String()
	for _, expected := range []string{"Status history", "Date", "Status", "new", "confirmed"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("page body does not contain %q: %s", expected, body)
		}
	}
}

func TestPaymentPageShowsExplicitStatusButtons(t *testing.T) {
	payments := newTestLibrePayment()

	id, err := payments.Register("merchant-1", 150.25, nil)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	handler := NewLibrePaymentHandler(payments, &PaymentUrlGenerator{baseURL: "http://localhost"})
	router := chi.NewRouter()
	router.Get("/librepayments/{payment_id}", handler.PaymentPage)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/librepayments/"+id, nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	body := recorder.Body.String()
	for _, expected := range []string{
		`<details class="librepayment-status-dropdown">`,
		`<summary>Set status</summary>`,
		`<div class="librepayment-status-menu">`,
		`class="librepayment-status-button"`,
		`librepaymentSetStatus(`,
		`'new'`,
		`'formShowed'`,
		`'deadlineExpired'`,
		`'canceled'`,
		`'authorizing'`,
		`'authorized'`,
		`'confirming'`,
		`'rejected'`,
		`'confirmed'`,
		`'refunding'`,
		`'refunded'`,
		`New`,
		`Form`,
		`Expired`,
		`Cancel`,
		`Auth`,
		`Authorized`,
		`Confirming`,
		`Rejected`,
		`Confirmed`,
		`Refund`,
		`Refunded`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("page body does not contain %q: %s", expected, body)
		}
	}
}

func TestIndexPageShowsExplicitStatusButtons(t *testing.T) {
	payments := newTestLibrePayment()

	_, err := payments.Register("merchant-1", 150.25, nil)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	handler := NewLibrePaymentHandler(payments, &PaymentUrlGenerator{baseURL: "http://localhost"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/librepayments", nil)
	handler.IndexPage(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	body := recorder.Body.String()
	for _, expected := range []string{
		`<details class="librepayment-status-dropdown">`,
		`<summary>Set status</summary>`,
		`<div class="librepayment-status-menu">`,
		`class="librepayment-status-button"`,
		`librepaymentSetStatus(`,
		`'new'`,
		`'formShowed'`,
		`'deadlineExpired'`,
		`'canceled'`,
		`'authorizing'`,
		`'authorized'`,
		`'confirming'`,
		`'rejected'`,
		`'confirmed'`,
		`'refunding'`,
		`'refunded'`,
		`New`,
		`Form`,
		`Expired`,
		`Cancel`,
		`Auth`,
		`Authorized`,
		`Confirming`,
		`Rejected`,
		`Confirmed`,
		`Refund`,
		`Refunded`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("index body does not contain %q: %s", expected, body)
		}
	}
}

func TestPaymentButtonsPassButtonElementToJS(t *testing.T) {
	payments := newTestLibrePayment()

	id, err := payments.Register("merchant-1", 150.25, nil)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	handler := NewLibrePaymentHandler(payments, &PaymentUrlGenerator{baseURL: "http://localhost"})
	router := chi.NewRouter()
	router.Get("/librepayments/{payment_id}", handler.PaymentPage)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/librepayments/"+id, nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	body := recorder.Body.String()
	for _, expected := range []string{
		`librepaymentCancel(this,`,
		`librepaymentConfirm(this,`,
		`librepaymentSetStatus(this,`,
		`.librepayment-status-dropdown`,
		`.librepayment-status-menu`,
		`librepayment-button-ok`,
		`librepayment-button-error`,
		`.librepayment-status-button:hover`,
		`cursor: pointer`,
		`padding:`,
		`border-radius:`,
		`transition:`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("page body does not contain %q: %s", expected, body)
		}
	}
}

func TestLibrePaymentJSUsesCancelEndpointForMainCancelAction(t *testing.T) {
	js, err := os.ReadFile("../../static/js/librepaymets.js")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	source := string(js)
	for _, expected := range []string{
		"function librepaymentCancel(button, payment_id)",
		"/cancel",
	} {
		if !strings.Contains(source, expected) {
			t.Fatalf("js source does not contain %q: %s", expected, source)
		}
	}

	if strings.Contains(source, "librepaymentReject") {
		t.Fatalf("main reject js method must be replaced by cancel: %s", source)
	}
}

func TestLibrePaymentJSHighlightsButtonByResponseStatus(t *testing.T) {
	js, err := os.ReadFile("../../static/js/librepaymets.js")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	source := string(js)
	for _, expected := range []string{
		"button.classList",
		"resp.status === 200",
		"librepayment-button-ok",
		"librepayment-button-error",
		"setTimeout",
		"1000",
		"window.location.reload();",
		".catch(",
	} {
		if !strings.Contains(source, expected) {
			t.Fatalf("js source does not contain %q: %s", expected, source)
		}
	}

	if strings.Contains(source, "setTimeout(() => {\n                window.location.reload();") {
		t.Fatalf("successful status change must reload without delay: %s", source)
	}
}

func newTestLibrePayment() *LibrePayment {
	return &LibrePayment{
		stor:        NewMemoryStorage(),
		notificator: NewNotificator(newNotificationJournal(), http.DefaultClient),
	}
}
