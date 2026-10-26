package deleter

import "strings"

// Reasons a delete can be blocked. The API error carries
// them as violation types, so a client can tell the reasons apart
// without reading the message text.
const (
	BlockerActiveSubscription   = "ACTIVE_SUBSCRIPTION"
	BlockerUnpaidInvoice        = "UNPAID_INVOICE"
	BlockerNegativeTokenBalance = "NEGATIVE_TOKEN_BALANCE"
	BlockerResourcesExist       = "RESOURCES_EXIST"
)

// Kinds of entity a Blocker's Subject id refers to.
const (
	SubjectSubscription      = "billing_subscription"
	SubjectInvoice           = "billing_invoice"
	SubjectBillingAccount    = "billing_account"
	SubjectResourceNamespace = "resource_namespace"
)

// Blocker is one reason a project or an organization cannot be deleted right now.
type Blocker struct {
	// Type is one of the Blocker* constants.
	Type string
	// Subject is the id of the entity behind the reason, or the namespace
	// for BlockerResourcesExist.
	Subject string
	// SubjectType is one of the Subject* constants and says what kind of
	// entity the Subject id refers to.
	SubjectType string
	// Count is the number of blocking resources, for BlockerResourcesExist.
	Count int
	// Message says what blocks the delete and how to clear it. Empty for
	// BlockerResourcesExist.
	Message string
}

// BlockedError carries every blocker the up-front check found, so the
// caller gets one checklist instead of discovering blockers one retry at
// a time.
type BlockedError struct {
	OrgID     string
	ProjectID string
	Blockers  []Blocker
}

func (e *BlockedError) Error() string {
	target := "organization"
	if e.ProjectID != "" {
		target = "project"
	}
	var msgs []string
	for _, b := range e.Blockers {
		if b.Message != "" {
			msgs = append(msgs, b.Message)
		}
	}
	if len(msgs) == 0 {
		return target + " cannot be deleted yet"
	}
	return target + " cannot be deleted yet: " + strings.Join(msgs, "; ")
}
