package tinkoff

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type StringInt string

func (st *StringInt) UnmarshalJSON(b []byte) error {
	var item interface{}
	if err := json.Unmarshal(b, &item); err != nil {
		return err
	}

	switch v := item.(type) {
	case string:
		*st = StringInt(v)
	case int:
		i := strconv.Itoa(v)
		*st = StringInt(i)
	case float64:
		f := fmt.Sprintf("%v", v)
		*st = StringInt(f)
	}

	return nil
}

type initInput struct {
	TerminalKey     string    `json:"TerminalKey"`
	Amount          int64     `json:"Amount"`
	OrderID         StringInt `json:"OrderId,number"`
	Description     string    `json:"Description"`
	Recurrent       string    `json:"Recurrent"`
	NotificationURL string    `json:"NotificationURL"`
	SuccessURL      string    `json:"SuccessURL"`
	FailURL         string    `json:"FailURL"`
	CustomerKey     StringInt `json:"CustomerKey"`
	Token           string    `json:"Token"`
}

type chargeInput struct {
	TerminalKey string `json:"TerminalKey"`
	ProcessID   int64  `json:"PaymentId"`
	RebillID    int64  `json:"RebillId"`
	Token       string `json:"Token"`
}

type getStateInput struct {
	TerminalKey string `json:"TerminalKey"`
	ProcessID   int64  `json:"PaymentId"`
	Token       string `json:"Token"`
}

func initRequest(r *http.Request) (input initInput, err error) {
	err = readJSON(r, &input)
	return input, err
}

func chargeRequest(r *http.Request) (input chargeInput, err error) {
	err = readJSON(r, &input)
	return input, err
}

func getStateRequest(r *http.Request) (input getStateInput, err error) {
	err = readJSON(r, &input)
	return input, err
}

func readJSON(r *http.Request, obj interface{}) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(obj)

	return err
}
