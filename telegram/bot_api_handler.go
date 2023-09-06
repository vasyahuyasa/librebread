package telegram

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

const contentTypeApplicationJson = "application/json"
const maxFormSize = 10 * 1024 * 1024

type BotAPI struct {
	storage MemoryTelegramStorage
	index   *indexPage
}

type response struct {
	Ok     bool        `json:"ok"`
	Result interface{} `json:"result"`
}

func NewBotAPI(storage *MemoryTelegramStorage) *BotAPI {
	return &BotAPI{
		storage: *storage,
		index:   newIndexPage(),
	}
}

func (api *BotAPI) IndexPage(w http.ResponseWriter, r *http.Request) {
	requests, err := api.storage.AllSortedByDateDesc()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("cannot get requests: %v", err)
		return
	}

	if isJson(r) {
		type JSONtelegramRequestEntry struct {
			Time    string            `json:"time"`
			Token   string            `json:"token"`
			Method  string            `json:"method"`
			Payload map[string]string `json:"payload"`
		}

		templateRequests := make([]JSONtelegramRequestEntry, 0, len(requests))

		for _, v := range requests {
			templateRequests = append(templateRequests, JSONtelegramRequestEntry{
				Time:    v.Time.Format(time.RFC3339),
				Token:   v.Token,
				Method:  v.Method,
				Payload: v.Payload,
			})
		}

		w.Header().Set("Content-Type", contentTypeApplicationJson)
		err = json.NewEncoder(w).Encode(templateRequests)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Printf("cannon encode json response")
			return
		}
	} else {

		templateRequests := make([]telegramRequestEntry, 0, len(requests))

		for _, v := range requests {
			t, err := request2TemplateEntry(v)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				log.Println(err)
				return
			}

			templateRequests = append(templateRequests, t)
		}
		err = api.index.writeTo(w, templateRequests)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}
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
			p := map[string]interface{}{}

			err := json.NewDecoder(r.Body).Decode(&p)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				log.Printf("cannot decode json request: %v", err)
				return
			}
			defer r.Body.Close()

			for k, v := range p {
				var strval string
				switch t := v.(type) {
				case float64:
					strval = strconv.FormatFloat(t, 'f', -1, 64)
				case string:
					strval = t
				}

				payload[k] = strval
			}
		} else {
			err := r.ParseMultipartForm(maxFormSize)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				log.Printf("cannot parse form: %v", err)
				return
			}

			for k, v := range r.PostForm {
				payload[k] = v[0]
			}
		}
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		log.Printf("method %q not allowed", r.Method)
		return
	}

	result, err := api.executeMethod(token, method, payload)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("cannot handle telegram method %q: %v", method, err)
	}

	w.Header().Set("Content-Type", contentTypeApplicationJson)
	err = json.NewEncoder(w).Encode(response{
		Ok:     true,
		Result: result,
	})
	if err != nil {
		log.Printf("cannot encode response %v", err)
	}
}

func (api *BotAPI) executeMethod(token string, method string, payload map[string]string) (interface{}, error) {
	p := make(map[string]string, len(payload))
	for k, v := range payload {
		p[k] = v
	}

	err := api.storage.Add(TelegramRequest{
		Time:    time.Now(),
		Token:   token,
		Method:  method,
		Payload: p,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot add telegram request to storage: %v", err)
	}
	switch method {
	case "sendMessage":
		return struct {
			MessageID int `json:"message_id"`
			From      struct {
				ID        int64  `json:"id"`
				IsBot     bool   `json:"is_bot"`
				FirstName string `json:"first_name"`
				Username  string `json:"username"`
			} `json:"from"`
			Chat struct {
				ID      int64  `json:"id"`
				Title   string `json:"title"`
				IsForum bool   `json:"is_forum"`
				Type    string `json:"type"`
			} `json:"chat"`
			Date            int `json:"date"`
			MessageThreadID int `json:"message_thread_id"`
			ReplyToMessage  struct {
				MessageID int `json:"message_id"`
				From      struct {
					ID        int    `json:"id"`
					IsBot     bool   `json:"is_bot"`
					FirstName string `json:"first_name"`
					LastName  string `json:"last_name"`
					Username  string `json:"username"`
				} `json:"from"`
				Chat struct {
					ID      int64  `json:"id"`
					Title   string `json:"title"`
					IsForum bool   `json:"is_forum"`
					Type    string `json:"type"`
				} `json:"chat"`
				Date              int `json:"date"`
				MessageThreadID   int `json:"message_thread_id"`
				ForumTopicCreated struct {
					Name      string `json:"name"`
					IconColor int    `json:"icon_color"`
				} `json:"forum_topic_created"`
				IsTopicMessage bool `json:"is_topic_message"`
			} `json:"reply_to_message"`
			Text     string `json:"text"`
			Entities []struct {
				Offset int    `json:"offset"`
				Length int    `json:"length"`
				Type   string `json:"type"`
			} `json:"entities"`
			IsTopicMessage bool `json:"is_topic_message"`
		}{
			MessageID:      123,
			Date:           int(time.Now().Unix()),
			Text:           payload["text"],
			IsTopicMessage: true,
		}, nil
	default:
		return struct{}{}, nil
	}
}

func isJson(r *http.Request) bool {
	s := strings.ToLower(r.FormValue("json"))

	return s == "1" || s == "true" || s == "yes"
}
