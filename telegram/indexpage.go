package telegram

import (
	_ "embed"
	"encoding/json"
	"io"
	"text/template"
	"time"
)

//go:embed index.tpl
var indexPageBlob string

type telegramRequestEntry struct {
	Time    string
	Token   string
	Method  string
	Payload string
}

type indexPage struct {
	indexPageTemplate *template.Template
}

func newIndexPage() *indexPage {
	return &indexPage{
		indexPageTemplate: template.Must(template.New("indexPage").Parse(indexPageBlob)),
	}
}

func (page *indexPage) writeTo(w io.Writer, entries []telegramRequestEntry) error {
	return page.indexPageTemplate.Execute(w, entries)
}

func request2TemplateEntry(req TelegramRequest) (telegramRequestEntry, error) {
	payloadAsJson, err := json.MarshalIndent(req.Payload, "", "  ")
	if err != nil {
		return telegramRequestEntry{}, err
	}

	return telegramRequestEntry{
		Time:    req.Time.Format(time.DateTime),
		Token:   req.Token,
		Method:  req.Method,
		Payload: string(payloadAsJson),
	}, nil
}
