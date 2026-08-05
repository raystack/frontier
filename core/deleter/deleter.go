package deleter

import "strings"

// Blocker types returned by the org delete pre-flight check. They are
// machine-readable and end up as PreconditionFailure violation types on
// the API error, so clients can branch on them.
const (
	BlockerActiveSubscription   = "ACTIVE_SUBSCRIPTION"
	BlockerUnpaidInvoice        = "UNPAID_INVOICE"
	BlockerNegativeTokenBalance = "NEGATIVE_TOKEN_BALANCE"
	BlockerUnusedTokens         = "UNUSED_TOKENS"
)

// Blocker is one reason an organization cannot be deleted right now. The
// message names the fix, and every fix is something the caller can do
// through the API themselves.
type Blocker struct {
	// Type is one of the Blocker* constants.
	Type string
	// Subject is the id of the blocking entity, e.g. a subscription id.
	Subject string
	// Message says what blocks the delete and what to do about it.
	Message string
}

// BlockedError carries every blocker the pre-flight check found, so the
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
