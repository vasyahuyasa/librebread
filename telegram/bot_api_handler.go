package telegram

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

const contentTypeApplicationJson = "application/json"

type BotAPI struct {
	storage MemoryTelegramStorage
}

type Response struct {
	Ok     bool        `json:"ok"`
	Result interface{} `json:"result"`
}

func (api *BotAPI) IndexPage(w http.ResponseWriter, r *http.Request) {

}

func (api *BotAPI) MethodHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "botToken")
	method := chi.URLParam(r, "botMethod")

	payload := map[string]string{}

	if r.Method == http.MethodGet {
		for k, v := range r.Form {
			payload[k] = v[0]
		}
	} else if r.Method == http.MethodPost {
		if r.Header.Get("Content-Type") == contentTypeApplicationJson {
			p := map[string]string{}

			err := json.NewDecoder(r.Body).Decode(&p)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				log.Printf("cannot decode json request: %v", err)
				return
			}
			defer r.Body.Close()

			for k, v := range p {
				payload[k] = v
			}
		} else {
			for k, v := range r.PostForm {
				payload[k] = v[0]
			}
		}
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		log.Printf("method %q not allowed", r.Method)
		return
	}

	err := api.executeMethod(token, method, payload)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("cannot handle telegram method %q: %v", method, err)
	}
}

func (api *BotAPI) executeMethod(token string, method string, payload map[string]string) error {
	p := make(map[string]string, len(payload))
	for k, v := range payload {
		p[k] = v
	}

	return api.storage.Add(TelegramRequest{
		Time:    time.Now(),
		Token:   token,
		Method:  method,
		Payload: p,
	})
}
