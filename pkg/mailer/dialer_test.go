package mailer

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/mail.v2"
)

// A server that accepts the connection but never sends its greeting must make
// the send fail after the dialer timeout instead of blocking forever.
func TestDialerImpl_DialAndSend_SilentServer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			defer conn.Close() // held open, never written to
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	d := NewDialerImpl("127.0.0.1", addr.Port, "", "", true, nil, mail.NoStartTLS)
	d.dialer.Timeout = 200 * time.Millisecond

	msg := mail.NewMessage()
	msg.SetHeader("From", "from@example.com")
	msg.SetHeader("To", "to@example.com")
	msg.SetBody("text/plain", "hi")

	done := make(chan error, 1)
	go func() { done <- d.DialAndSend(msg) }()

	select {
	case err := <-done:
		assert.Error(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("DialAndSend still blocked after 3s: no deadline on the greeting read")
	}
}
