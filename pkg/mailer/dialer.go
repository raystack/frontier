package mailer

import (
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/raystack/frontier/pkg/mailer/mocks"
	"github.com/stretchr/testify/mock"

	"gopkg.in/mail.v2"
)

const (
	defaultFromHeader = "frontier@raystack.org"
)

var setDialOnce sync.Once

// installGreetingDeadline makes every SMTP connection get a deadline as soon
// as it opens. gopkg.in/mail.v2 only sets the deadline after it has read the
// server greeting, so a server that accepts the connection and then stays
// silent would block the send forever.
func installGreetingDeadline() {
	setDialOnce.Do(func() {
		dial := mail.NetDialTimeout
		mail.NetDialTimeout = func(network, address string, timeout time.Duration) (net.Conn, error) {
			conn, err := dial(network, address, timeout)
			if err != nil {
				return nil, err
			}
			if timeout > 0 {
				if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
					conn.Close()
					return nil, err
				}
			}
			return conn, nil
		}
	})
}

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
	installGreetingDeadline()
	d := mail.NewDialer(SMTPHost, SMTPPort, SMTPUser, SMTPPass)
	d.TLSConfig = &tls.Config{
		InsecureSkipVerify: SMTPInsecure, //nolint:gosec
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
