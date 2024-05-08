package strategy

import (
	"bytes"
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

func mockDialer(t *testing.T, email string, otp string) *mocks.Dialer {
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
		wantLen         int
		wantErr         bool
		testUsersConfig testusers.Config
	}{
		{
			name: "should send mail using random chars",
			fields: fields{
				dialer:  mockDialer(t, "test@acme.org", "******"),
				subject: "auth otp",
				body:    "here is the otp, use it: {{.Otp}}",
			},
			args: args{
				to: "test@acme.org",
			},
			wantLen:         6,
			wantErr:         false,
			testUsersConfig: testusers.Config{},
		},
		{
			name: "should send mail for test user",
			fields: fields{
				dialer:  mockDialer(t, "abc@acme2.org", "111111"),
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
			if tt.want != "" && got != tt.want {
				t.Errorf("SendMail() got = %v, want %v", got, tt.want)
			}
			if tt.wantLen > 0 && len(got) != tt.wantLen {
				t.Errorf("SendMail() got = %v, want %v", len(got), tt.wantLen)
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
		if compareLine(line, gotLines[i]) {
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

// compareLine compare two strings ignoring characters that has '*' in place of them
func compareLine(l, r string) bool {
	if len(l) != len(r) {
		return false
	}
	for i := 0; i < len(l); i++ {
		if l[i] != '*' && l[i] != r[i] {
			return false
		}
	}
	return true
}
