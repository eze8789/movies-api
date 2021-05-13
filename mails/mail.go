package mails

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	dialer *mail.Dialer
	sender string
}

func New(h string, p int, user, password, sender string) Mailer {
	dialer := mail.NewDialer(h, p, user, password)
	dialer.Timeout = 5 * time.Second

	return Mailer{dialer: dialer,
		sender: sender,
	}
}

func (m *Mailer) Send(email, templateFile string, params interface{}) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", params)
	if err != nil {
		return err
	}

	pBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(pBody, "plainBody", params)
	if err != nil {
		return err
	}

	hBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(hBody, "htmlBody", params)
	if err != nil {
		return err
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", email)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", hBody.String())
	msg.AddAlternative("text/html", hBody.String())

	// retry email send if fails, the process is in a different goroutine
	// so no UX affected
	for i := 0; i <= 3; i++ {
		err = m.dialer.DialAndSend(msg)
		// return if ok
		if err == nil {
			return nil
		}
		time.Sleep(2 * time.Second) //nolint:gomnd
	}
	return err
}
