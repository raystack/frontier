package mailer

import (
	"bufio"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/mail.v2"
)

func greetingTestMessage() *mail.Message {
	msg := mail.NewMessage()
	msg.SetHeader("From", "frontier@raystack.org")
	msg.SetHeader("To", "user@example.com")
	msg.SetHeader("Subject", "greeting deadline test")
	msg.SetBody("text/plain", "hello")
	return msg
}

// A server that accepts the connection but never sends its greeting must
// make the send fail with a timeout instead of blocking forever: mail.v2
// reads the greeting before it sets any connection deadline.
func TestDialerImpl_DialAndSend_SilentServer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	accepted := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			accepted <- conn // held open and never written to
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	impl := NewDialerImpl(addr.IP.String(), addr.Port, "", "", true, nil, mail.NoStartTLS)
	impl.dialer.Timeout = 300 * time.Millisecond

	done := make(chan error, 1)
	go func() {
		done <- impl.DialAndSend(greetingTestMessage())
	}()

	select {
	case err := <-done:
		require.Error(t, err)
		require.True(t, os.IsTimeout(err), "expected a timeout error, got: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("DialAndSend still blocked after 10s: no deadline on the greeting read")
	}

	select {
	case conn := <-accepted:
		conn.Close()
	default:
	}
}

// A server that answers normally must still complete the send: the
// greeting deadline must not break regular delivery.
func TestDialerImpl_DialAndSend_CooperativeServer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	go serveFakeSMTP(listener)

	addr := listener.Addr().(*net.TCPAddr)
	impl := NewDialerImpl(addr.IP.String(), addr.Port, "", "", true, nil, mail.NoStartTLS)

	done := make(chan error, 1)
	go func() {
		done <- impl.DialAndSend(greetingTestMessage())
	}()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("DialAndSend did not complete against a cooperative server")
	}
}

// serveFakeSMTP speaks just enough SMTP for one successful delivery.
func serveFakeSMTP(listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	write := func(s string) { _, _ = conn.Write([]byte(s)) }
	reader := bufio.NewReader(conn)
	write("220 fake ESMTP\r\n")
	inData := false
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		if inData {
			if strings.TrimRight(line, "\r\n") == "." {
				inData = false
				write("250 OK\r\n")
			}
			continue
		}
		switch {
		case strings.HasPrefix(line, "EHLO"), strings.HasPrefix(line, "HELO"):
			write("250 fake\r\n")
		case strings.HasPrefix(line, "DATA"):
			inData = true
			write("354 go ahead\r\n")
		case strings.HasPrefix(line, "QUIT"):
			write("221 Bye\r\n")
			return
		default:
			write("250 OK\r\n")
		}
	}
}
