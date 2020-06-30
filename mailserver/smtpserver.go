package mailserver

import (
	"log"
	"time"

	"github.com/emersion/go-smtp"
)

type SmtpServer struct {
	addr     string
	stor     *MailStorage
	notifier EmailNotifier
}

func NewSmtpServer(addr string, stor *MailStorage, notifier EmailNotifier) *SmtpServer {
	return &SmtpServer{
		addr:     addr,
		stor:     stor,
		notifier: notifier,
	}
}

func (srv *SmtpServer) ListenAndServe() error {
	be := &backend{
		store:    srv.stor,
		notifier: srv.notifier,
	}

	s := smtp.NewServer(be)

	s.Addr = srv.addr
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	s.AllowInsecureAuth = true

	log.Println("start smtp on", s.Addr)
	return s.ListenAndServe()
}
