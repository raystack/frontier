package deleter

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/core/deleter/mocks"
	pkgauditrecord "github.com/raystack/frontier/pkg/auditrecord"
	mailermocks "github.com/raystack/frontier/pkg/mailer/mocks"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/mail.v2"
)

func deletedRecord(accountID string, meta metadata.Metadata) auditrecord.AuditRecord {
	return auditrecord.AuditRecord{
		Event:    pkgauditrecord.BillingCustomerDeletedEvent,
		Resource: auditrecord.Resource{ID: accountID, Type: pkgauditrecord.BillingCustomerType},
		OrgID:    "org-1",
		Metadata: meta,
	}
}

func recoverWith(t *testing.T, records []auditrecord.AuditRecord, listErr error, notice deleteNotice) deleteNotice {
	t.Helper()
	auditSvc := mocks.NewAuditRecordService(t)
	auditSvc.EXPECT().List(mock.Anything, mock.Anything).
		Return(auditrecord.AuditRecordsList{AuditRecords: records}, listErr)
	Service{auditRecordService: auditSvc}.recoverForfeitFromAudit(context.Background(), "org-1", &notice)
	return notice
}

func TestRecoverForfeitFromAudit(t *testing.T) {
	forfeit := metadata.Metadata{customer.DeleteRecordForfeitedTokensKey: float64(70), customer.DeleteRecordPurchasedTokensKey: float64(30), customer.DeleteRecordOrgTeardownKey: true}

	t.Run("adds the forfeit of an account that is gone", func(t *testing.T) {
		got := recoverWith(t, []auditrecord.AuditRecord{deletedRecord("cust-1", forfeit)}, nil, deleteNotice{})
		assert.Equal(t, int64(70), got.Amount)
		assert.Equal(t, int64(30), got.Purchased)
	})

	t.Run("skips an account deleted on its own", func(t *testing.T) {
		standalone := metadata.Metadata{customer.DeleteRecordForfeitedTokensKey: float64(70), customer.DeleteRecordPurchasedTokensKey: float64(30)}
		got := recoverWith(t, []auditrecord.AuditRecord{deletedRecord("cust-1", standalone)}, nil, deleteNotice{})
		assert.Equal(t, deleteNotice{}, got)
	})

	t.Run("skips a record whose teardown flag is not a boolean", func(t *testing.T) {
		malformed := metadata.Metadata{customer.DeleteRecordForfeitedTokensKey: float64(70), customer.DeleteRecordOrgTeardownKey: "true"}
		got := recoverWith(t, []auditrecord.AuditRecord{deletedRecord("cust-1", malformed)}, nil, deleteNotice{})
		assert.Equal(t, deleteNotice{}, got)
	})

	t.Run("skips an account that is still live", func(t *testing.T) {
		live := deleteNotice{
			Amount:    100,
			Purchased: 100,
			Accounts:  map[string]struct{}{"cust-1": {}},
		}
		got := recoverWith(t, []auditrecord.AuditRecord{deletedRecord("cust-1", forfeit)}, nil, live)
		assert.Equal(t, live, got)
	})

	t.Run("leaves the notice unchanged when the audit records cannot be listed", func(t *testing.T) {
		start := deleteNotice{Amount: 5, Purchased: 2}
		got := recoverWith(t, []auditrecord.AuditRecord{deletedRecord("cust-1", forfeit)}, errors.New("db down"), start)
		assert.Equal(t, start, got)
	})

	t.Run("ignores a record with no resource id", func(t *testing.T) {
		got := recoverWith(t, []auditrecord.AuditRecord{deletedRecord("", forfeit)}, nil, deleteNotice{})
		assert.Equal(t, deleteNotice{}, got)
	})
}

func TestParseMailTemplate(t *testing.T) {
	render := func(t *testing.T, tmpl mailTemplate) (string, string) {
		t.Helper()
		var subject, body bytes.Buffer
		assert.NoError(t, tmpl.subject.Execute(&subject, nil))
		assert.NoError(t, tmpl.body.Execute(&body, nil))
		return subject.String(), body.String()
	}

	t.Run("an empty subject and body fall back to the defaults", func(t *testing.T) {
		tmpl, err := parseMailTemplate("", "", "default subject", "default body")
		assert.NoError(t, err)
		subject, body := render(t, tmpl)
		assert.Equal(t, "default subject", subject)
		assert.Equal(t, "default body", body)
	})

	t.Run("a configured subject and body win over the defaults", func(t *testing.T) {
		tmpl, err := parseMailTemplate("set subject", "set body", "default subject", "default body")
		assert.NoError(t, err)
		subject, body := render(t, tmpl)
		assert.Equal(t, "set subject", subject)
		assert.Equal(t, "set body", body)
	})

	t.Run("a broken subject fails", func(t *testing.T) {
		_, err := parseMailTemplate("{{.Name", "", "", "default body")
		assert.ErrorContains(t, err, "subject")
	})

	t.Run("a broken body fails", func(t *testing.T) {
		_, err := parseMailTemplate("", "{{.Name", "default subject", "")
		assert.ErrorContains(t, err, "body")
	})
}

func TestSendMail(t *testing.T) {
	tmpl, err := parseMailTemplate("Hi {{.Name}}", "<b>{{.Name}}</b>", "", "")
	assert.NoError(t, err)

	t.Run("sends one mail to the address", func(t *testing.T) {
		dialer := mailermocks.NewDialer(t)
		dialer.EXPECT().FromHeader().Return("no-reply@frontier.test")
		dialer.EXPECT().DialAndSend(mock.Anything).Run(func(msg *mail.Message) {
			assert.Equal(t, []string{"no-reply@frontier.test"}, msg.GetHeader("From"))
			assert.Equal(t, []string{"maya@example.com"}, msg.GetHeader("To"))
			assert.Equal(t, []string{"Hi Maya"}, msg.GetHeader("Subject"))
		}).Return(nil)

		err := Service{mailDialer: dialer}.sendMail(tmpl, "maya@example.com", struct{ Name string }{Name: "Maya"})
		assert.NoError(t, err)
	})

	t.Run("a template that fails to render sends nothing", func(t *testing.T) {
		dialer := mailermocks.NewDialer(t)

		err := Service{mailDialer: dialer}.sendMail(tmpl, "maya@example.com", struct{}{})
		assert.ErrorContains(t, err, "subject")
	})
}
