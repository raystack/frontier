package deleter

import (
	"slices"
	"strings"

	"github.com/raystack/frontier/core/organization"
)

// Reasons a delete can be blocked. The API error carries
// them as violation types, so a client can tell the reasons apart
// without reading the message text.
const (
	BlockerActiveSubscription   = "ACTIVE_SUBSCRIPTION"
	BlockerUnpaidInvoice        = "UNPAID_INVOICE"
	BlockerNegativeTokenBalance = "NEGATIVE_TOKEN_BALANCE"
	BlockerResourcesExist       = "RESOURCES_EXIST"
	BlockerSoleOwnerWithMembers = "SOLE_OWNER_WITH_MEMBERS"
)

// Kinds of entity a Blocker's Subject id refers to.
const (
	SubjectSubscription      = "billing_subscription"
	SubjectInvoice           = "billing_invoice"
	SubjectBillingAccount    = "billing_account"
	SubjectResourceNamespace = "resource_namespace"
)

// Blocker is one reason a project, an organization or a user cannot be deleted right now.
type Blocker struct {
	// Type is one of the Blocker* constants.
	Type string
	// Subject is the id of the entity behind the reason, or the namespace
	// for BlockerResourcesExist.
	Subject string
	// SubjectType is one of the Subject* constants, or the organization
	// namespace, and says what kind of entity the Subject id refers to.
	SubjectType string
	// Count is the number of blocking resources, for BlockerResourcesExist.
	Count int
	// Message says what blocks the delete and how to clear it. Empty for
	// BlockerResourcesExist.
	Message string
}

type UserDeleteCheck struct {
	Blockers              []Blocker
	OrganizationsToDelete []OrganizationToDelete
	OrganizationsToLeave  []OrganizationToLeave
	PATCount              int
}

type OrganizationToDelete struct {
	Organization organization.Organization
	Blockers     []Blocker
}

type OrganizationToLeave struct {
	Organization organization.Organization
	RoleTitle    string
	Blockers     []Blocker
}

func (c UserDeleteCheck) CanDelete() bool {
	return len(c.allBlockers()) == 0
}

func (c UserDeleteCheck) allBlockers() []Blocker {
	blockers := slices.Clone(c.Blockers)
	for _, o := range c.OrganizationsToDelete {
		blockers = append(blockers, o.Blockers...)
	}
	for _, o := range c.OrganizationsToLeave {
		blockers = append(blockers, o.Blockers...)
	}
	return blockers
}

// BlockedError carries every blocker the up-front check found, so the
// caller gets one checklist instead of discovering blockers one retry at
// a time.
type BlockedError struct {
	OrgID     string
	ProjectID string
	UserID    string
	Blockers  []Blocker
}

func (e *BlockedError) Error() string {
	target := "organization"
	if e.ProjectID != "" {
		target = "project"
	}
	if e.UserID != "" {
		target = "user"
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
