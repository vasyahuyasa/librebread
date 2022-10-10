package librepayment

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "embed"

	"github.com/go-chi/chi/v5"
)

var (
	//go:embed index.tpl
	indexPageData string

	//go:embed payment.tpl
	paymentPageData string
)

type LibrePaymentHandler struct {
	p            *LibrePayment
	urlGenerator *PaymentUrlGenerator

	indexPageTemplate   *template.Template
	paymentPageTemplate *template.Template
}

func NewLibrePaymentHandler(p *LibrePayment, urlGenerator *PaymentUrlGenerator) *LibrePaymentHandler {

	return &LibrePaymentHandler{
		p:                   p,
		urlGenerator:        urlGenerator,
		indexPageTemplate:   template.Must(template.New("indexPage").Parse(indexPageData)),
		paymentPageTemplate: template.Must(template.New("paymentPage").Parse(paymentPageData)),
	}
}

func (h *LibrePaymentHandler) RegisterPayment(w http.ResponseWriter, r *http.Request) {
	amountStr := r.FormValue("amount")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse %q as float64: %v", amountStr, err), http.StatusBadRequest)
		return
	}

	merchant := r.FormValue("merchant")

	payload := map[string]string{}

	for key, values := range r.Form {
		if key == "amount" || key == "merchant" {
			continue
		}

		payload[key] = strings.Join(values, ",")
	}

	id, err := h.p.Register(merchant, amount, payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type response struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}

	err = json.NewEncoder(w).Encode(response{
		ID:  id,
		URL: h.urlGenerator.Generate(id),
	})

	if err != nil {
		log.Printf("cannot send reponse: %v", err)
	}
}

func (h *LibrePaymentHandler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "payment_id")

	payment, err := h.p.Status(id)
	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, ErrPaymentNotFound) {
			code = http.StatusNotFound
		}

		http.Error(w, err.Error(), code)
		return
	}

	type response struct {
		CreatedAT string  `json:"created_at"`
		ID        string  `json:"id"`
		Status    string  `json:"status"`
		Amount    float64 `json:"amount"`
		Merchant  string  `json:"merchant"`
	}

	err = json.NewEncoder(w).Encode(response{
		CreatedAT: payment.CreatedAt.Format("2006-01-02 15:04:05"),
		ID:        payment.ID,
		Status:    payment.Status,
		Amount:    payment.Amount,
		Merchant:  payment.Merchant,
	})

	if err != nil {
		log.Printf("cannot encode json response: %v", err)
	}
}

func (h *LibrePaymentHandler) ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "payment_id")

	err := h.p.stor.SetPaymentStatus(id, StatusConfirmed)
	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, ErrPaymentNotFound) {
			code = http.StatusNotFound
		}

		http.Error(w, err.Error(), code)
		return
	}
}

func (h *LibrePaymentHandler) RejectPayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "payment_id")

	err := h.p.stor.SetPaymentStatus(id, StatusRejected)
	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, ErrPaymentNotFound) {
			code = http.StatusNotFound
		}

		http.Error(w, err.Error(), code)
		return
	}
}

func (h *LibrePaymentHandler) IndexPage(w http.ResponseWriter, r *http.Request) {
	type templatePayment struct {
		Time     string
		ID       string
		Amount   float64
		Merchant string
		Status   string
	}

	payments, err := h.p.AllPaymentsDescOrder()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templatepayments := make([]templatePayment, len(payments))
	for i, p := range payments {
		templatepayments[i] = templatePayment{
			Time:     p.CreatedAt.Format("2006-01-02 15:04:05"),
			ID:       p.ID,
			Amount:   p.Amount,
			Merchant: p.Merchant,
			Status:   p.Status,
		}
	}

	err = h.indexPageTemplate.Execute(w, templatepayments)
	if err != nil {
		log.Printf("cannot execute template: %v", err)
	}
}

func (h *LibrePaymentHandler) PaymentPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "payment_id")

	payment, err := h.p.Status(id)
	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, ErrPaymentNotFound) {
			code = http.StatusNotFound
		}

		http.Error(w, err.Error(), code)
		return
	}

	type templatePayment struct {
		Time     string
		ID       string
		Amount   float64
		Merchant string
		Status   string

		Payload map[string]string
	}

	templateData := templatePayment{
		Time:     payment.CreatedAt.Format("2006-01-02 15:04:05"),
		ID:       payment.ID,
		Amount:   payment.Amount,
		Merchant: payment.Merchant,
		Status:   payment.Status,

		Payload: map[string]string{},
	}

	for k, v := range payment.Payload {
		templateData.Payload[k] = v
	}

	err = h.paymentPageTemplate.Execute(w, templateData)
	if err != nil {
		log.Printf("cannot execute template: %v", err)
	}
}
