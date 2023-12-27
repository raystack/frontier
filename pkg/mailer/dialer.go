package mailer

import (
	"crypto/tls"

	"github.com/raystack/frontier/pkg/mailer/mocks"
	"github.com/stretchr/testify/mock"

	"gopkg.in/mail.v2"
)

const (
	defaultFromHeader = "frontier@raystack.org"
)

type Dialer interface {
	DialAndSend(m *mail.Message) error
	FromHeader() string
}

func NewMockDialer() *mocks.Dialer {
	d := &mocks.Dialer{}
	d.EXPECT().DialAndSend(mock.Anything).Return(nil)
	d.EXPECT().FromHeader().Return(defaultFromHeader)
	return d
}

type DialerImpl struct {
	dialer  *mail.Dialer
	headers map[string]string
}

func NewDialerImpl(SMTPHost string, SMTPPort int, SMTPUser string, SMTPPass string,
	SMTPInsecure bool, headers map[string]string, tlsPolicy mail.StartTLSPolicy) *DialerImpl {
	d := mail.NewDialer(SMTPHost, SMTPPort, SMTPUser, SMTPPass)
	d.TLSConfig = &tls.Config{
		InsecureSkipVerify: SMTPInsecure,
		ServerName:         SMTPHost,
	}
	d.StartTLSPolicy = tlsPolicy
	return &DialerImpl{
		dialer:  d,
		headers: headers,
	}
}

// FromHeader returns the headers to be added to the mail as from field
func (m DialerImpl) FromHeader() string {
	if _, ok := m.headers["from"]; !ok {
		return defaultFromHeader
	}
	return m.headers["from"]
}

func (m DialerImpl) DialAndSend(msg *mail.Message) error {
	return m.dialer.DialAndSend(msg)
}
