package deleter

import (
	"context"
	"errors"
	"testing"

	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/core/deleter/mocks"
	pkgauditrecord "github.com/raystack/frontier/pkg/auditrecord"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	forfeit := metadata.Metadata{"forfeited_tokens": float64(70), "purchased_tokens": float64(30)}

	t.Run("adds the forfeit of an account that is gone", func(t *testing.T) {
		got := recoverWith(t, []auditrecord.AuditRecord{deletedRecord("cust-1", forfeit)}, nil, deleteNotice{})
		assert.Equal(t, int64(70), got.Amount)
		assert.Equal(t, int64(30), got.Purchased)
	})

	t.Run("skips an account that is still live", func(t *testing.T) {
		live := deleteNotice{
			Amount:    100,
			Purchased: 100,
			Accounts:  map[string]accountTokens{"cust-1": {Balance: 100, Purchased: 100}},
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
