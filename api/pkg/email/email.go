package email

import (
	"fmt"
	"net/smtp"
)

type Sender interface {
	Send(to []string, subject, body string) error
}

type Mailer struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewMailer(host, port, username, password, from string) *Mailer {
	return &Mailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (m *Mailer) Send(to []string, subject, body string) error {
	auth := smtp.PlainAuth("", m.username, m.password, m.host)

	header := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n",
		m.from, joinAddresses(to), subject,
	)

	addr := fmt.Sprintf("%s:%s", m.host, m.port)
	return smtp.SendMail(addr, auth, m.from, to, []byte(header+body))
}

func joinAddresses(addrs []string) string {
	result := ""
	for i, a := range addrs {
		if i > 0 {
			result += ", "
		}
		result += a
	}
	return result
}
