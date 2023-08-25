package telegram

import "net/http"

type BotAPI struct{}

func (api *BotAPI) Handler(w http.ResponseWriter, r *http.Request) {}

func (api *BotAPI) handleMethod(token string, method string, payload map[string]string) {

}
