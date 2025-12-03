package email

import (
	"fmt"
	"net/smtp"
)

type Sender struct {
	host string
	port int
	from string
}

func NewSender(host string, port int, from string) *Sender {
	return &Sender{
		host: host,
		port: port,
		from: from,
	}
}

func (s *Sender) SendVerification(to, token, baseURL string) error {
	link := fmt.Sprintf("%s/auth/verify?token=%s", baseURL, token)

	subject := "Подтвердите вашу регистрацию"
	body := fmt.Sprintf(`Здравствуйте!

Для подтверждения регистрации перейдите по ссылке:
%s

Если вы не регистрировались, просто проигнорируйте это письмо.

С уважением,
GoMarket Team`, link)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		s.from, to, subject, body)

	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	return smtp.SendMail(addr, nil, s.from, []string{to}, []byte(msg))
}
