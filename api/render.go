package api

import (
	"html/template"
	"log"
	"net/http"
)

type renderer struct {
	smsTemplate          *template.Template
	helpdeskeddyTemplate *template.Template
}

func newRenderer() *renderer {
	r := &renderer{}

	r.smsTemplate = parseTemplates("sms", baseTmplate, smsTemplate)
	r.helpdeskeddyTemplate = parseTemplates("helpdeskeddy", baseTmplate, helpdeskeddyTemplate)

	return r
}

func parseTemplates(name string, tmpls ...string) *template.Template {
	t := template.New(name)
	var err error

	for _, s := range tmpls {
		t, err = t.Parse(s)
		if err != nil {
			log.Fatalf("can not parse template %s: %v", name, err)
		}
	}

	return t
}

func (re *renderer) renderSms(w http.ResponseWriter, smses SMSList) error {
	return re.smsTemplate.Execute(w, smses)
}
