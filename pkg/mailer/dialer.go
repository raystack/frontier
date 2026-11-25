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

var greetingDeadlineOnce sync.Once

// installGreetingDeadline wraps the mail library's dial hook so that a
// deadline is in force from the moment a connection opens.
//
// mail.v2 reads the server greeting inside Dial before it sets the
// connection deadline from Dialer.Timeout, so a server that accepts the
// connection but never sends a greeting blocks that read forever. Setting
// the deadline inside the dial hook covers the greeting read too; the
// library re-arms the same deadline once the greeting arrives, so the
// rest of the session behaves as before. The deadline is derived from
// the timeout the library passes to the hook for each dial, so every
// dialer keeps its own Timeout, and a timeout of zero keeps the
// library's unbounded behaviour.
func installGreetingDeadline() {
	greetingDeadlineOnce.Do(func() {
		dial := mail.NetDialTimeout
		mail.NetDialTimeout = func(network, addr string, timeout time.Duration) (net.Conn, error) {
			conn, err := dial(network, addr, timeout)
			if err != nil || timeout <= 0 {
				return conn, err
			}
			if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
				conn.Close()
				return nil, err
			}
			return conn, nil
		}
	})
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
