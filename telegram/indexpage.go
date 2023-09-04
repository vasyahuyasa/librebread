package telegram

import (
	_ "embed"
	"text/template"
)

type IndexPage struct {
	indexPageTemplate *template.Template
}

func NewIndexPage() *IndexPage {
	//go:embed index.tpl
	var page string

	return &IndexPage{
		indexPageTemplate: template.Must(template.New("indexPage").Parse(page)),
	}
}
