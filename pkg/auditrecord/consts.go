package auditrecord

// Event represents an audit event type
type Event string

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
	OrganizationCreateEvent      Event = "organization.create"
	OrganizationUpdateEvent      Event = "organization.update"
	OrganizationStateChangeEvent Event = "organization.state_change"
	OrganizationDeleteEvent      Event = "organization.delete"
	OrganizationInvitedEvent     Event = "organization.invited"

	// KYC Events
	KYCVerifiedEvent   Event = "kyc.verified"
	KYCUnverifiedEvent Event = "kyc.unverified"
	SystemActor              = "system"
)

// String returns the string representation of the event
func (e Event) String() string {
	return string(e)
}
