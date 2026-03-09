package auditrecord

// Event represents an audit event type
type Event string

// EntityType represents the type of a resource or target in an audit record
type EntityType string

const (
	// Billing Customer Events
	BillingCustomerCreatedEvent       Event = "billing_customer.created"
	BillingCustomerUpdatedEvent       Event = "billing_customer.updated"
	BillingCustomerCreditUpdatedEvent Event = "billing_customer.credit_updated"
	BillingCustomerDeletedEvent       Event = "billing_customer.deleted"

	// Billing Checkout Events
	BillingCheckoutCreatedEvent Event = "billing_checkout.created"

	// Billing Subscription Events
	BillingSubscriptionCreatedEvent Event = "billing_subscription.created"
	BillingSubscriptionChangedEvent Event = "billing_subscription.changed"

	// Billing Transaction Events
	BillingTransactionDebitEvent  Event = "billing_transaction.debit"
	BillingTransactionCreditEvent Event = "billing_transaction.credit"

	// Service User Events
	ServiceUserCreatedEvent Event = "serviceuser.created"
	ServiceUserDeletedEvent Event = "serviceuser.deleted"

	// Organization Events
	OrganizationCreateEvent             Event = "organization.create"
	OrganizationUpdateEvent             Event = "organization.update"
	OrganizationStateChangeEvent        Event = "organization.state_change"
	OrganizationDeleteEvent             Event = "organization.delete"
	OrganizationInvitedEvent            Event = "organization.invited"
	OrganizationMemberAddedEvent        Event = "organization.added"
	OrganizationMemberRemovedEvent      Event = "organization.removed"
	OrganizationInvitationAcceptedEvent Event = "organization.accepted"

	// KYC Events
	KYCVerifiedEvent   Event = "kyc.verified"
	KYCUnverifiedEvent Event = "kyc.unverified"

	// Role Events
	RoleCreatedEvent Event = "role.created"
	RoleUpdatedEvent Event = "role.updated"

	// Policy Events
	PolicyCreatedEvent Event = "policy.created"
	PolicyUpdatedEvent Event = "policy.updated"
	PolicyDeletedEvent Event = "policy.deleted"

	// Session Events
	SessionRevokedEvent Event = "session.revoked"

	// PAT Events
	PATCreatedEvent Event = "pat.created"

	SystemActor = "system"

	// Entity Types (used in Resource.Type and Target.Type)
	OrganizationType        EntityType = "organization"
	ProjectType             EntityType = "project"
	GroupType               EntityType = "group"
	UserType                EntityType = "user"
	RoleType                EntityType = "role"
	PolicyType              EntityType = "policy"
	ServiceUserType         EntityType = "serviceuser"
	InvitationType          EntityType = "invitation"
	KycType                 EntityType = "kyc"
	BillingCustomerType     EntityType = "billing_customer"
	BillingCheckoutType     EntityType = "billing_checkout"
	BillingSubscriptionType EntityType = "billing_subscription"
	BillingTransactionType  EntityType = "billing_transaction"
	SessionType             EntityType = "session"
	PATType                 EntityType = "pat"
)

// String returns the string representation of the event
func (e Event) String() string {
	return string(e)
}

// String returns the string representation of the entity type
func (t EntityType) String() string {
	return string(t)
}
