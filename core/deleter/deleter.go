package deleter

import "strings"

// Reasons an organization delete can be blocked. The API error carries
// them as violation types, so a client can tell the reasons apart
// without reading the message text.
const (
	BlockerActiveSubscription   = "ACTIVE_SUBSCRIPTION"
	BlockerUnpaidInvoice        = "UNPAID_INVOICE"
	BlockerNegativeTokenBalance = "NEGATIVE_TOKEN_BALANCE"
)

// Kinds of entity a Blocker's Subject id refers to.
const (
	SubjectSubscription   = "billing_subscription"
	SubjectInvoice        = "billing_invoice"
	SubjectBillingAccount = "billing_account"
)

// Blocker is one reason an organization cannot be deleted right now.
type Blocker struct {
	// Type is one of the Blocker* constants.
	Type string
	// Subject is the id of the entity behind the reason.
	Subject string
	// SubjectType is one of the Subject* constants and says what kind of
	// entity the Subject id refers to.
	SubjectType string
	// Message says what blocks the delete and how to clear it.
	Message string
}

// BlockedError carries every blocker the up-front check found, so the
// caller gets one checklist instead of discovering blockers one retry at
// a time.
type BlockedError struct {
	OrgID    string
	Blockers []Blocker
}

func (e *BlockedError) Error() string {
	msgs := make([]string, 0, len(e.Blockers))
	for _, b := range e.Blockers {
		msgs = append(msgs, b.Message)
	}
	return "organization cannot be deleted yet: " + strings.Join(msgs, "; ")
}
