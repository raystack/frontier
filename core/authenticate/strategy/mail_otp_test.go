package strategy

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"
	"time"

	testusers "github.com/raystack/frontier/core/authenticate/test_users"
	"github.com/raystack/frontier/pkg/mailer"
	"github.com/raystack/frontier/pkg/mailer/mocks"
	"github.com/stretchr/testify/mock"
	"gopkg.in/mail.v2"
)

var mockDate = time.Date(2023, 10, 6, 0, 0, 0, 0, time.UTC)

func mock1(t *testing.T, email string, otp string) *mocks.Dialer {
	t.Helper()

	wantMsg := "MIME-Version: 1.0\r\n" +
		"From: frontier@acme.org\r\n" +
		"To: " + email + "\r\n" +
		"Subject: auth otp\r\n" +
		"Date: " + mockDate.Format(time.RFC1123Z) + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: quoted-printable\r\n" +
		"\r\n" +
		`here is the otp, use it: ` + otp

	mockDialer1 := &mocks.Dialer{}
	mockDialer1.EXPECT().FromHeader().Return("frontier@acme.org")
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
	type fields struct {
		dialer  mailer.Dialer
		subject string
		body    string
	}
	type args struct {
		to string
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		want            string
		wantErr         bool
		testUsersConfig testusers.Config
	}{
		{
			name: "should send mail",
			fields: fields{
				dialer:  mock1(t, "test@acme.org", "7GAPMQ"),
				subject: "auth otp",
				body:    "here is the otp, use it: {{.Otp}}",
			},
			args: args{
				to: "test@acme.org",
			},
			want:            "7GAPMQ",
			wantErr:         false,
			testUsersConfig: testusers.Config{},
		},
		{
			name: "should send mail",
			fields: fields{
				dialer:  mock1(t, "test@acme1.org", "7GAPMQ"),
				subject: "auth otp",
				body:    "here is the otp, use it: {{.Otp}}",
			},
			args: args{
				to: "test@acme1.org",
			},
			want:    "7GAPMQ",
			wantErr: false,
			testUsersConfig: testusers.Config{
				Enabled: true,
				Domain:  "acme2.org",
				OTP:     "111111",
			},
		},
		{
			name: "should send mail",
			fields: fields{
				dialer:  mock1(t, "abc@acme2.org", "111111"),
				subject: "auth otp",
				body:    "here is the otp, use it: {{.Otp}}",
			},
			args: args{
				to: "abc@acme2.org",
			},
			want:    "111111",
			wantErr: false,
			testUsersConfig: testusers.Config{
				Enabled: true,
				Domain:  "acme2.org",
				OTP:     "111111",
			},
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
			got, err := m.SendMail(tt.args.to, tt.testUsersConfig)
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
