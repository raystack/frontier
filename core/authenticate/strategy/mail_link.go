package strategy

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/raystack/frontier/pkg/crypt"

	testusers "github.com/raystack/frontier/core/authenticate/test_users"
	"github.com/raystack/frontier/pkg/mailer"
	"github.com/raystack/frontier/pkg/utils"
	"gopkg.in/mail.v2"
)

const (
	MailLinkAuthMethod string = "maillink"
)

// MailLink sends a mail with a one time password link to user's email id.
// On successful verification, it creates a session
type MailLink struct {
	dialer  mailer.Dialer
	subject string
	body    string
	Now     func() time.Time
	host    string
}

func NewMailLink(d mailer.Dialer, host, subject, body string) *MailLink {
	return &MailLink{
		host:    host,
		dialer:  d,
		subject: subject,
		body:    body,
		Now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// SendMail sends a mail with a one time password embedded link to user's email id
func (m MailLink) SendMail(id, to string, testUsersConfig testusers.Config) (string, error) {
	var otp string
	userDomain := utils.ExtractDomainFromEmail(to)
	if testUsersConfig.Enabled && userDomain == testUsersConfig.Domain && len(testUsersConfig.OTP) > 0 {
		otp = testUsersConfig.OTP
	} else {
		otp = crypt.GenerateRandomStringFromLetters(otpLen, otpLetterRunes)
	}

	t, err := template.New("body").Parse(m.body)
	if err != nil {
		return "", fmt.Errorf("failed to parse email template: %w", err)
	}
	var tpl bytes.Buffer

	link := fmt.Sprintf("%s?strategy_name=%s&code=%s&state=%s", strings.TrimRight(m.host, "/"), MailLinkAuthMethod, otp, id)
	err = t.Execute(&tpl, map[string]string{
		"Link": link,
	})
	if err != nil {
		return "", fmt.Errorf("failed to parse email template: %w", err)
	}

	// TODO(kushsharma): apply rest of the headers
	msg := mail.NewMessage()
	msg.SetHeader("From", m.dialer.FromHeader())
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", m.subject)
	msg.SetBody("text/html", tpl.String())
	msg.SetDateHeader("Date", m.Now())
	return otp, m.dialer.DialAndSend(msg)
}
