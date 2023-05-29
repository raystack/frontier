package strategy

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/odpf/shield/pkg/mailer"
	"github.com/odpf/shield/pkg/mailer/mocks"
	"github.com/stretchr/testify/mock"
	"gopkg.in/mail.v2"
)

var mockDate = time.Date(2023, 10, 6, 0, 0, 0, 0, time.UTC)

func mock1(t *testing.T) *mocks.Dialer {
	t.Helper()

	wantMsg := "MIME-Version: 1.0\r\n" +
		"From: shield@acme.org\r\n" +
		"To: test@acme.org\r\n" +
		"Subject: auth otp\r\n" +
		"Date: " + mockDate.Format(time.RFC1123Z) + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: quoted-printable\r\n" +
		"\r\n" +
		`here is the otp, use it: bnVqZ3`

	mockDialer1 := &mocks.Dialer{}
	mockDialer1.EXPECT().FromHeader().Return("shield@acme.org")
	mockDialer1.On("DialAndSend", mock.MatchedBy(func(m *mail.Message) bool {
		buf := new(bytes.Buffer)
		_, err := m.WriteTo(buf)
		if err != nil {
			return false
		}
		return compareBodies(t, buf.String(), wantMsg)
	})).Return(nil)
	return mockDialer1
}

func TestMailOTP_SendMail(t *testing.T) {
	mockDialer1 := mock1(t)
	defer mockDialer1.AssertExpectations(t)

	type fields struct {
		dialer  mailer.Dialer
		subject string
		body    string
	}
	type args struct {
		to string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "should send mail",
			fields: fields{
				dialer:  mockDialer1,
				subject: "auth otp",
				body:    "here is the otp, use it: {{.Otp}}",
			},
			args: args{
				to: "test@acme.org",
			},
			want:    "bnVqZ3",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rand.Seed(1)
			m := MailOTP{
				dialer:  tt.fields.dialer,
				subject: tt.fields.subject,
				body:    tt.fields.body,
				Now: func() time.Time {
					return mockDate
				},
			}
			got, err := m.SendMail(tt.args.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendMail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SendMail() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func compareBodies(t *testing.T, got, want string) bool {
	t.Helper()
	// We cannot do a simple comparison since the ordering of headers' fields
	// is random.
	gotLines := strings.Split(got, "\r\n")
	wantLines := strings.Split(want, "\r\n")

	// We only test for too many lines, missing lines are tested after
	if len(gotLines) > len(wantLines) {
		t.Fatalf("Message has too many lines, \ngot %d:\n%s\nwant %d:\n%s", len(gotLines), got, len(wantLines), want)
		return false
	}

	isInHeader := true
	headerStart := 0
	for i, line := range wantLines {
		if line == gotLines[i] {
			if line == "" {
				isInHeader = false
			} else if !isInHeader && len(line) > 2 && line[:2] == "--" {
				isInHeader = true
				headerStart = i + 1
			}
			continue
		}

		if !isInHeader {
			missingLine(t, line, got, want)
			return false
		}

		isMissing := true
		for j := headerStart; j < len(gotLines); j++ {
			if gotLines[j] == "" {
				break
			}
			if gotLines[j] == line {
				isMissing = false
				break
			}
		}
		if isMissing {
			missingLine(t, line, got, want)
			return false
		}
	}

	return true
}

func missingLine(t *testing.T, line, got, want string) {
	t.Helper()
	t.Fatalf("Missing line %q\ngot:\n%s\nwant:\n%s", line, got, want)
}
