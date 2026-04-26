package mailer

import (
	"fmt"
	"net/smtp"

	"github.com/vidwadeseram/go-auth-template/internal/config"
)

type Mailer struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) Send(recipient string, subject string, body string) error {
	message := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", recipient, subject, body))
	return smtp.SendMail(m.cfg.MailAddress(), nil, m.cfg.MailFrom, []string{recipient}, message)
}
