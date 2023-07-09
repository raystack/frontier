package strategy

import (
	"bytes"
	"fmt"
	"html/template"
	"math/rand"
	"time"

	"github.com/raystack/shield/pkg/mailer"
	"gopkg.in/mail.v2"
)

var (
	otpLetterRunes = []rune("ABCDEFGHJKMNPQRSTWXYZ23456789")
	otpLen         = 6
)

// MailOTP sends a mail with a one time password to user's email id
// and verifies the OTP. On successful verification, it creates a session
type MailOTP struct {
	dialer  mailer.Dialer
	subject string
	body    string
	Now     func() time.Time
}

func NewMailLink(d mailer.Dialer, subject, body string) *MailOTP {
	return &MailOTP{
		dialer:  d,
		subject: subject,
		body:    body,
		Now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// SendMail sends a mail with a one time password embedded link to user's email id
func (m MailOTP) SendMail(to string) (string, error) {
	otp := GenerateNonceFromLetters(otpLen, otpLetterRunes)
	t, err := template.New("body").Parse(m.body)
	if err != nil {
		return "", fmt.Errorf("failed to parse email template: %w", err)
	}
	var tpl bytes.Buffer
	err = t.Execute(&tpl, map[string]string{
		"Otp": otp,
	})
	if err != nil {
		return "", fmt.Errorf("failed to parse email template: %w", err)
	}

	//TODO(kushsharma): apply rest of the headers
	msg := mail.NewMessage()
	msg.SetHeader("From", m.dialer.FromHeader())
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", m.subject)
	msg.SetBody("text/html", tpl.String())
	msg.SetDateHeader("Date", m.Now())
	return otp, m.dialer.DialAndSend(msg)
}

func GenerateNonceFromLetters(length int, letterRunes []rune) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
