package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi"
	"github.com/vasyahuyasa/librebread/helpdesk"
	"github.com/vasyahuyasa/librebread/mailserver"
	"github.com/vasyahuyasa/librebread/sms"
	"github.com/vasyahuyasa/librebread/ssenotifier"
)

const (
	TLSaddr     = ":443"
	addr        = ":80"
	smtpAddr    = ":25"
	pop3Addr    = ":110"
	filename    = "messages.txt"
	helpdekFile = "helpdesk.msgp"
	emailFile   = "email.msgp"

	tplHeader = `
	<html>
		<head>
			<style>
				ol {
					padding: 10px; 
					list-style-type: none;				
				}
				ol li {
					float: left;
					margin: 0 10px 0 0;
				}
			</style>
		</head>
		<body>
			<ol>
				<li><a href="/">sms</a></li>
				<li><a href="/helpdesk">helpdesk</a></li>
				<li><a href="/email">email</a></li>
			</ol>`

	smsTableFooter = `</table>`

	helpdeskTableFooter = `</table>`

	emailTableFooter = `</table>`

	tplFooter = `</body></html>`
)

func helpdeskTableHeaderWithCount(feedbackCount int) string {
	const helpdeskTableHeader = `
	<table border=1>
		<caption>Helpdesk (%d)</caption>
		<thead>
			<th>Date</th>
			<th>Title</th>
			<th>Description</th>
		</thead>`

	return fmt.Sprintf(helpdeskTableHeader, feedbackCount)
}

func smsTableHeaderWithCount(messageCount int) string {
	const smsTableHeader = `
	<table border=1>
	    <caption>SMS (%d)</caption>
		<thead>
			<th>Date</th>
			<th>From</th>
			<th>Phone</th>
			<th>Msg</th>
			<th>Provider</th>
		</thead>`

	return fmt.Sprintf(smsTableHeader, messageCount)
}

func emailTableHeaderWithCount(count int) string {

	return fmt.Sprintf(`
	<table border=1>
		<caption>Email (%d)</caption>
		<thead>
			<th>Time</th>
			<th>From</th>
			<th>To</th>
			<th>Subject</th>
			<th>Data</th>
		</thead>`, count)
}

func main() {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Fatal("can not open file:", err)
	}
	defer f.Close()

	smsStor := sms.NewStorage(f)

	err = smsStor.Restore()
	if err != nil {
		log.Fatal("can not restore SMS messages:", err)
	}

	hdf, err := os.OpenFile(helpdekFile, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Fatal("can not open helpdesk file:", err)
	}
	defer hdf.Close()

	hstor := helpdesk.NewStorage(hdf)

	err = hstor.Restore()
	if err != nil {
		log.Fatal("can not restore HelpDesk messages:", err)
	}

	mf, err := os.OpenFile(emailFile, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Fatal("can not open email file:", err)
	}

	mailStor := mailserver.NewStorage(mf)

	err = mailStor.Restore()
	if err != nil {
		log.Fatal("can not restore email messages:", err)
	}

	sseNotifier := ssenotifier.NewBroker()

	smsru := sms.SmsRu{
		Stor:     smsStor,
		Notifier: sseNotifier,
	}

	devino := sms.Devino{
		Stor:     smsStor,
		Notifier: sseNotifier,
	}

	// smtp
	go func() {
		smtpsrv := mailserver.NewSmtpServer(smtpAddr, mailStor, sseNotifier)

		err := smtpsrv.ListenAndServe()
		if err != nil {
			log.Fatalf("smtp server failed: %v", err)
		}
	}()

	// pop3
	go func() {
		pop3 := mailserver.NewPopServer(pop3Addr)
		err := pop3.ListenAndServe()
		if err != nil {
			log.Fatalf("pop3 server failed: %v", err)
		}
	}()

	go func() {
		httpServer(smsStor, hstor, smsru, mailStor, sseNotifier)
	}()

	// devino telecom mock server
	r := chi.NewRouter()
	r.Use(caselessMatcher)

	devinoTelecomRoutes(r, devino)
	smsRuRoutes(r, smsru)
	helpdeskRoutes(r, hstor, sseNotifier)

	log.Println("start HTTPS on", TLSaddr)
	err = http.ListenAndServeTLS(TLSaddr, "cert/server.crt", "cert/server.key", r)
	if err != nil {
		log.Println("TLS Web server fail:", err)
	}
}

func devinoTelecomRoutes(r chi.Router, devino sms.Devino) {
	r.Route("/rest", func(r chi.Router) {
		r.Post("/user/sessionid", devino.UserSessionIdHandler)
		r.Post("/sms/send", devino.SmsSend)
		r.Post("/sms/state", devino.SmsState)
	})

	r.Route("/rest/v2", func(r chi.Router) {
		r.Post("/sms/send", devino.SmsSend)
		r.Post("/sms/state", devino.SmsState)
	})
}

func smsRuRoutes(mux *chi.Mux, smsru sms.SmsRu) {
	mux.Route("/sms", func(r chi.Router) {
		r.Post("/send", smsru.Send)
		r.Post("/status", smsru.Status)
	})
}

func helpdeskRoutes(mux *chi.Mux, stor *helpdesk.HelpdeskStorage, notifier helpdesk.HelpdeskNotifier) {
	mux.Post("/api/v2/tickets/", helpdesk.HelpdeskEddyHandler(stor, notifier))
}

// sms.ru and stats server
func httpServer(stor *sms.Storage, hstor *helpdesk.HelpdeskStorage, smsru sms.SmsRu, mailStor *mailserver.MailStorage, sseNotification *ssenotifier.Broker) {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(indexPageWrapper)
		r.Get("/", indexSmsHandler(stor))
		r.Get("/helpdesk", helpdeskIndexHandler(hstor))
		r.Get("/email", emailIndexHandler(mailStor))
	})

	r.Get("/events", sseNotification.ClientHandler())

	smsRuRoutes(r, smsru)
	helpdeskRoutes(r, hstor, sseNotification)

	log.Println("start HTTP on", addr)
	err := http.ListenAndServe(addr, r)
	if err != nil {
		log.Println("Web server fail:", err)
	}
}

func indexSmsHandler(stor *sms.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b := strings.Builder{}
		b.WriteString(smsTableHeaderWithCount(stor.Len()))
		for _, msg := range stor.LastMessages(50) {
			b.WriteString("<tr>" +
				"<td>" + msg.Time.Format("2006-01-02 15:04:05") + "</td>" +
				"<td>" + msg.From + "</td>" +
				"<td>" + msg.To + "</td>" +
				"<td>" + msg.Text + "</td>" +
				"<td>" + msg.Provider + "</td>" +
				"</tr>")
		}
		b.WriteString(smsTableFooter)
		_, err := w.Write([]byte(b.String()))
		if err != nil {
			log.Printf("can not send index to client: %v", err)
		}
	}
}

func helpdeskIndexHandler(stor *helpdesk.HelpdeskStorage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b := strings.Builder{}
		b.WriteString(helpdeskTableHeaderWithCount(stor.Len()))
		for _, msg := range stor.LastMessages(50) {
			b.WriteString("<tr>" +
				"<td>" + msg.Time.Format("2006-01-02 15:04:05") + "</td>" +
				"<td>" + msg.Title + "</td>" +
				"<td>" + msg.Description + "</td>" +
				"</tr>")
		}
		b.WriteString(helpdeskTableFooter)
		_, err := w.Write([]byte(b.String()))
		if err != nil {
			log.Printf("can not send index to client: %v", err)
		}
	}
}

func emailIndexHandler(stor *mailserver.MailStorage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b := strings.Builder{}
		b.WriteString(emailTableHeaderWithCount(stor.Len()))
		for _, msg := range stor.LastMessages() {
			b.WriteString("<tr>" +
				"<td>" + msg.SentOn.Format("2006-01-02 15:04:05") + "</td>" +
				"<td>" + html.EscapeString(msg.From) + "</td>" +
				"<td>" + html.EscapeString(msg.To) + "</td>" +
				"<td>" + html.EscapeString(msg.Subject) + "</td>" +
				"<td>" + html.EscapeString(msg.Body) + "</td>" +
				"</tr>")
		}
		b.WriteString(emailTableFooter)
		_, err := w.Write([]byte(b.String()))
		if err != nil {
			log.Printf("can not send index to client: %v", err)
		}
	}
}

func indexPageWrapper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, tplHeader)
		next.ServeHTTP(w, r)
		fmt.Fprint(w, tplFooter)
	})
}

// caselessMatcher is convert request path to lowercase
func caselessMatcher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.ToLower(r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
