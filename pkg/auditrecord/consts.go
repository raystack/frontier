package auditrecord

// Event represents an audit event type
type Event string

// Type represents the type of a resource or target in an audit record
type Type string

const (
	// Billing Customer Events
	BillingCustomerCreatedEvent       Event = "billing_customer.created"
	BillingCustomerUpdatedEvent       Event = "billing_customer.updated"
	BillingCustomerCreditUpdatedEvent Event = "billing_customer.credit_updated"
	BillingCustomerDeletedEvent       Event = "billing_customer.deleted"

	// Billing Checkout Events
	BillingCheckoutCreatedEvent Event = "billing_checkout.created"

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

	SystemActor = "system"

	// Entity Types (used in Resource.Type and Target.Type)
	OrganizationType       Type = "organization"
	UserType               Type = "user"
	RoleType               Type = "role"
	ServiceUserType        Type = "serviceuser"
	InvitationType         Type = "invitation"
	KycType                Type = "kyc"
	BillingCustomerType    Type = "billing_customer"
	BillingCheckoutType    Type = "billing_checkout"
	BillingTransactionType Type = "billing_transaction"
)

// String returns the string representation of the event
func (e Event) String() string {
	return string(e)
}

// String returns the string representation of the type
func (t Type) String() string {
	return string(t)
}
