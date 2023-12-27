package strategy

import (
	"bytes"
	"fmt"
	"html/template"
	"math/rand"
	"time"

	"github.com/raystack/frontier/pkg/mailer"
	"gopkg.in/mail.v2"
)

const (
	MailOTPAuthMethod string = "mailotp"
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

func NewMailOTP(d mailer.Dialer, subject, body string) *MailOTP {
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

	tpl := template.New("body")
	t, err := tpl.Parse(m.body)
	if err != nil {
		return "", fmt.Errorf("failed to parse email template: %w", err)
	}
	var tplBuffer bytes.Buffer
	if err = t.Execute(&tplBuffer, map[string]string{
		"Otp": otp,
	}); err != nil {
		return "", fmt.Errorf("failed to parse email template: %w", err)
	}
	tplBody := tplBuffer.String()

	tpl = template.New("sub")
	t, err = tpl.Parse(m.subject)
	if err != nil {
		return "", fmt.Errorf("failed to parse email template: %w", err)
	}
	tplBuffer.Reset()
	if err = t.Execute(&tplBuffer, map[string]string{
		"Otp": otp,
	}); err != nil {
		return "", fmt.Errorf("failed to parse email template: %w", err)
	}
	tplSub := tplBuffer.String()

	//TODO(kushsharma): apply rest of the headers
	msg := mail.NewMessage()
	msg.SetHeader("From", m.dialer.FromHeader())
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", tplSub)
	msg.SetBody("text/html", tplBody)
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
