package push

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type LibreBreadHandler struct {
	librePush *LibrePush
}

func NewLibreBreadHandler(librePush *LibrePush) *LibreBreadHandler {
	return &LibreBreadHandler{
		librePush: librePush,
	}
}

func (h *LibreBreadHandler) HandlePush(w http.ResponseWriter, r *http.Request) {
	var msg struct {
		ID           int64             `json:"id"`
		PushService  string            `json:"push_service"`
		Title        string            `json:"title"`
		Text         string            `json:"text"`
		Data         map[string]string `json:"data,omitempty"`
		TTL          int64             `json:"ttl"`
		Tokens       []string          `json:"tokens"`
		ValidateOnly bool              `json:"validate_only"`
	}

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&msg)
	if err != nil {
		log.Printf("cannot decode request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tokens := msg.Tokens
	pushService := msg.PushService
	msg.Tokens = nil

	msgAsBytes, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		log.Printf("cannot marshal request as json: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response *BatchResponse

	if msg.ValidateOnly {
		response, err = h.dryRun(tokens)
	} else {
		response, err = h.send(pushService, msgAsBytes, tokens)
	}

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("cannot encode json reponse: %v", err)
	}
}

func (h *LibreBreadHandler) HandlePush2(w http.ResponseWriter, r *http.Request) {
	var reqData struct {
		PushService  string          `json:"push_source"`
		RawMsg       json.RawMessage `json:"raw_msg"`
		Tokens       []string        `json:"tokens"`
		ValidateOnly bool            `json:"validate_only"`
	}

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&reqData)
	if err != nil {
		log.Printf("cannot decode request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var response *BatchResponse

	if reqData.ValidateOnly {
		response, err = h.dryRun(reqData.Tokens)
	} else {
		response, err = h.send(reqData.PushService, reqData.RawMsg, reqData.Tokens)
	}

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("cannot encode json reponse: %v", err)
	}
}

func (h *LibreBreadHandler) send(pushService string, rawMsg []byte, tokens []string) (*BatchResponse, error) {
	response, err := h.librePush.Send(pushService, rawMsg, tokens)
	if err != nil {
		return nil, fmt.Errorf("cannot emulate push send: %w", err)
	}

	return response, nil
}

func (h *LibreBreadHandler) dryRun(tokens []string) (*BatchResponse, error) {
	response, err := h.librePush.SendDryRun(tokens)
	if err != nil {
		return nil, fmt.Errorf("cannot dry run push send: %w", err)
	}

	return response, nil
}
