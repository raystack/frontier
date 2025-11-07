package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/raystack/frontier/billing/subscription"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
	"github.com/raystack/frontier/pkg/db"
)

type SubscriptionChanges struct {
	Phases      []Phase `json:"phases"`
	PlanHistory []Phase `json:"plan_history"`
}

type Phase struct {
	EffectiveAt time.Time `json:"effective_at"`
	EndsAt      time.Time `json:"ends_at"`
	PlanID      string    `json:"plan_id"`
	Reason      string    `json:"reason"`
}

func (c *SubscriptionChanges) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return json.Unmarshal(src, c)
	case string:
		return json.Unmarshal([]byte(src), c)
	case nil:
		return nil
	}
	return fmt.Errorf("cannot convert %T to JsonB", src)
}

func (c SubscriptionChanges) Value() (driver.Value, error) {
	return json.Marshal(c)
}

type Subscription struct {
	ID             string `db:"id"`
	ProviderID     string `db:"provider_id"`
	SubscriptionID string `db:"customer_id"`
	PlanID         string `db:"plan_id"`

	State    string              `db:"state"`
	Metadata types.NullJSONText  `db:"metadata"`
	Changes  SubscriptionChanges `db:"changes"`

	CreatedAt            time.Time  `db:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at"`
	CanceledAt           *time.Time `db:"canceled_at"`
	EndedAt              *time.Time `db:"ended_at"`
	DeletedAt            *time.Time `db:"deleted_at"`
	TrialEndsAt          *time.Time `db:"trial_ends_at"`
	CurrentPeriodStartAt *time.Time `db:"current_period_start_at"`
	CurrentPeriodEndAt   *time.Time `db:"current_period_end_at"`
	BillingCycleAnchorAt *time.Time `db:"billing_cycle_anchor_at"`
}

func (c Subscription) transform() (subscription.Subscription, error) {
	var unmarshalledMetadata map[string]any
	if c.Metadata.Valid {
		if err := c.Metadata.Unmarshal(&unmarshalledMetadata); err != nil {
			return subscription.Subscription{}, err
		}
	}
	canceledAt := time.Time{}
	if c.CanceledAt != nil {
		canceledAt = *c.CanceledAt
	}
	endedAt := time.Time{}
	if c.EndedAt != nil {
		endedAt = *c.EndedAt
	}
	deletedAt := time.Time{}
	if c.DeletedAt != nil {
		deletedAt = *c.DeletedAt
	}
	trialEndsAt := time.Time{}
	if c.TrialEndsAt != nil {
		trialEndsAt = *c.TrialEndsAt
	}
	currentPeriodStartAt := time.Time{}
	if c.CurrentPeriodStartAt != nil {
		currentPeriodStartAt = *c.CurrentPeriodStartAt
	}
	currentPeriodEndAt := time.Time{}
	if c.CurrentPeriodEndAt != nil {
		currentPeriodEndAt = *c.CurrentPeriodEndAt
	}
	billingCycleAnchorAt := time.Time{}
	if c.BillingCycleAnchorAt != nil {
		billingCycleAnchorAt = *c.BillingCycleAnchorAt
	}
	phase := subscription.Phase{}
	if len(c.Changes.Phases) > 0 {
		// we only care about the first change at the moment
		phase = subscription.Phase(c.Changes.Phases[0])
	}
	planHistory := make([]subscription.Phase, 0, len(c.Changes.PlanHistory))
	for _, p := range c.Changes.PlanHistory {
		planHistory = append(planHistory, subscription.Phase(p))
	}
	return subscription.Subscription{
		ID:                   c.ID,
		ProviderID:           c.ProviderID,
		CustomerID:           c.SubscriptionID,
		PlanID:               c.PlanID,
		State:                c.State,
		Metadata:             unmarshalledMetadata,
		Phase:                phase,
		PlanHistory:          planHistory,
		CreatedAt:            c.CreatedAt,
		CanceledAt:           canceledAt,
		UpdatedAt:            c.UpdatedAt,
		DeletedAt:            deletedAt,
		EndedAt:              endedAt,
		TrialEndsAt:          trialEndsAt,
		CurrentPeriodStartAt: currentPeriodStartAt,
		CurrentPeriodEndAt:   currentPeriodEndAt,
		BillingCycleAnchorAt: billingCycleAnchorAt,
	}, nil
}

// subscriptionWithCustomer extends Subscription to include customer info for audit logging
type subscriptionWithCustomer struct {
	Subscription
	CustomerOrgID string `db:"customer_org_id"`
	CustomerName  string `db:"customer_name"`
}

type BillingSubscriptionRepository struct {
	dbc *db.Client
}

func NewBillingSubscriptionRepository(dbc *db.Client) *BillingSubscriptionRepository {
	return &BillingSubscriptionRepository{
		dbc: dbc,
	}
}

// buildCustomerSubqueries creates subqueries to fetch customer org_id and name
func (r BillingSubscriptionRepository) buildCustomerSubqueries() (orgIDSubquery, nameSubquery *goqu.SelectDataset) {
	orgIDSubquery = dialect.From(TABLE_BILLING_CUSTOMERS).
		Select("org_id").
		Where(goqu.Ex{"id": goqu.I(TABLE_BILLING_SUBSCRIPTIONS + ".customer_id")})

	nameSubquery = dialect.From(TABLE_BILLING_CUSTOMERS).
		Select("name").
		Where(goqu.Ex{"id": goqu.I(TABLE_BILLING_SUBSCRIPTIONS + ".customer_id")})

	return orgIDSubquery, nameSubquery
}

func (r BillingSubscriptionRepository) Create(ctx context.Context, toCreate subscription.Subscription) (subscription.Subscription, error) {
	if toCreate.Metadata == nil {
		toCreate.Metadata = make(map[string]any)
	}
	marshaledMetadata, err := json.Marshal(toCreate.Metadata)
	if err != nil {
		return subscription.Subscription{}, err
	}
	if toCreate.ID == "" {
		toCreate.ID = uuid.New().String()
	}

	record := goqu.Record{
		"id":          toCreate.ID,
		"provider_id": toCreate.ProviderID,
		"customer_id": toCreate.CustomerID,
		"plan_id":     toCreate.PlanID,
		"metadata":    marshaledMetadata,
		"updated_at":  goqu.L("now()"),
	}
	if !toCreate.TrialEndsAt.IsZero() {
		record["trial_ends_at"] = toCreate.TrialEndsAt
	}
	if !toCreate.CurrentPeriodStartAt.IsZero() {
		record["current_period_start_at"] = toCreate.CurrentPeriodStartAt
	}
	if !toCreate.CurrentPeriodEndAt.IsZero() {
		record["current_period_end_at"] = toCreate.CurrentPeriodEndAt
	}
	if !toCreate.BillingCycleAnchorAt.IsZero() {
		record["billing_cycle_anchor_at"] = toCreate.BillingCycleAnchorAt
	}
	if toCreate.State != "" {
		record["state"] = toCreate.State
	}

	// Fetch customer info for audit
	customerOrgIDSubquery, customerNameSubquery := r.buildCustomerSubqueries()

	query, params, err := dialect.Insert(TABLE_BILLING_SUBSCRIPTIONS).
		Rows(record).Returning(
		goqu.I(TABLE_BILLING_SUBSCRIPTIONS+".*"),
		customerOrgIDSubquery.As("customer_org_id"),
		customerNameSubquery.As("customer_name"),
	).ToSQL()
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var result subscriptionWithCustomer
	if err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_BILLING_SUBSCRIPTIONS, "Create", func(ctx context.Context) error {
			if err := tx.QueryRowxContext(ctx, query, params...).StructScan(&result); err != nil {
				return err
			}

			// Create audit record for subscription creation
			return r.createSubscriptionAuditRecord(
				ctx,
				tx,
				pkgAuditRecord.BillingSubscriptionCreatedEvent,
				result,
				map[string]interface{}{
					"plan_id": result.PlanID,
					"state":   result.State,
				},
				result.CreatedAt,
			)
		})
	}); err != nil {
		return subscription.Subscription{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return result.Subscription.transform()
}

func (r BillingSubscriptionRepository) GetByID(ctx context.Context, id string) (subscription.Subscription, error) {
	stmt := dialect.Select().From(TABLE_BILLING_SUBSCRIPTIONS).Where(goqu.Ex{
		"id": id,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var subscriptionModel Subscription
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_SUBSCRIPTIONS, "GetByID", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&subscriptionModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return subscription.Subscription{}, subscription.ErrNotFound
		}
		return subscription.Subscription{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return subscriptionModel.transform()
}

func (r BillingSubscriptionRepository) GetByName(ctx context.Context, name string) (subscription.Subscription, error) {
	stmt := dialect.Select().From(TABLE_BILLING_SUBSCRIPTIONS).Where(goqu.Ex{
		"name": name,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var subscriptionModel Subscription
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_SUBSCRIPTIONS, "GetByName", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&subscriptionModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return subscription.Subscription{}, subscription.ErrNotFound
		}
		return subscription.Subscription{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return subscriptionModel.transform()
}

func (r BillingSubscriptionRepository) GetByProviderID(ctx context.Context, id string) (subscription.Subscription, error) {
	stmt := dialect.Select().From(TABLE_BILLING_SUBSCRIPTIONS).Where(goqu.Ex{
		"provider_id": id,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var subscriptionModel Subscription
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_SUBSCRIPTIONS, "GetByName", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&subscriptionModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return subscription.Subscription{}, subscription.ErrNotFound
		}
		return subscription.Subscription{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return subscriptionModel.transform()
}

func (r BillingSubscriptionRepository) UpdateByID(ctx context.Context, toUpdate subscription.Subscription) (subscription.Subscription, error) {
	if strings.TrimSpace(toUpdate.ID) == "" {
		return subscription.Subscription{}, subscription.ErrInvalidID
	}
	if strings.TrimSpace(toUpdate.State) == "" {
		return subscription.Subscription{}, subscription.ErrInvalidDetail
	}

	// Get old subscription to check for plan change
	oldSub, err := r.GetByID(ctx, toUpdate.ID)
	if err != nil {
		return subscription.Subscription{}, err
	}

	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	updateRecord := goqu.Record{
		"metadata":   marshaledMetadata,
		"updated_at": goqu.L("now()"),
	}
	if toUpdate.PlanID != "" {
		updateRecord["plan_id"] = toUpdate.PlanID
	}
	if toUpdate.State != "" {
		updateRecord["state"] = toUpdate.State
	}
	if !toUpdate.CanceledAt.IsZero() {
		updateRecord["canceled_at"] = toUpdate.CanceledAt
	}
	if !toUpdate.EndedAt.IsZero() {
		updateRecord["ended_at"] = toUpdate.EndedAt
	}
	if !toUpdate.TrialEndsAt.IsZero() {
		updateRecord["trial_ends_at"] = toUpdate.TrialEndsAt
	}
	if !toUpdate.CurrentPeriodStartAt.IsZero() {
		updateRecord["current_period_start_at"] = toUpdate.CurrentPeriodStartAt
	}
	if !toUpdate.CurrentPeriodEndAt.IsZero() {
		updateRecord["current_period_end_at"] = toUpdate.CurrentPeriodEndAt
	}
	if !toUpdate.BillingCycleAnchorAt.IsZero() {
		updateRecord["billing_cycle_anchor_at"] = toUpdate.BillingCycleAnchorAt
	}
	updateRecord["changes"] = r.toSubscriptionChanges(toUpdate)

	// Fetch customer info for audit
	customerOrgIDSubquery, customerNameSubquery := r.buildCustomerSubqueries()

	query, params, err := dialect.Update(TABLE_BILLING_SUBSCRIPTIONS).Set(updateRecord).Where(goqu.Ex{
		"id": toUpdate.ID,
	}).Returning(
		goqu.I(TABLE_BILLING_SUBSCRIPTIONS+".*"),
		customerOrgIDSubquery.As("customer_org_id"),
		customerNameSubquery.As("customer_name"),
	).ToSQL()
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var result subscriptionWithCustomer
	if err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_BILLING_SUBSCRIPTIONS, "Update", func(ctx context.Context) error {
			if err := tx.QueryRowxContext(ctx, query, params...).StructScan(&result); err != nil {
				err = checkPostgresError(err)
				switch {
				case errors.Is(err, sql.ErrNoRows):
					return subscription.ErrNotFound
				case errors.Is(err, ErrInvalidTextRepresentation):
					return subscription.ErrInvalidUUID
				default:
					return err
				}
			}

			// Audit if plan changed
			if oldSub.PlanID != result.PlanID {
				return r.createSubscriptionAuditRecord(
					ctx,
					tx,
					pkgAuditRecord.BillingSubscriptionChangedEvent,
					result,
					map[string]interface{}{
						"old_plan_id": oldSub.PlanID,
						"new_plan_id": result.PlanID,
					},
					result.UpdatedAt,
				)
			}

			return nil
		})
	}); err != nil {
		return subscription.Subscription{}, fmt.Errorf("%w: %w", txnErr, err)
	}

	return result.Subscription.transform()
}

func (r BillingSubscriptionRepository) toSubscriptionChanges(toUpdate subscription.Subscription) SubscriptionChanges {
	subChanges := SubscriptionChanges{
		Phases: []Phase{
			Phase(toUpdate.Phase),
		},
	}
	planHistory := make([]Phase, 0, len(toUpdate.PlanHistory))
	for _, p := range toUpdate.PlanHistory {
		planHistory = append(planHistory, Phase(p))
	}
	subChanges.PlanHistory = planHistory
	return subChanges
}

func (r BillingSubscriptionRepository) List(ctx context.Context, filter subscription.Filter) ([]subscription.Subscription, error) {
	stmt := dialect.Select().From(TABLE_BILLING_SUBSCRIPTIONS).Order(goqu.I("created_at").Desc())
	if filter.CustomerID != "" {
		stmt = stmt.Where(goqu.Ex{
			"customer_id": filter.CustomerID,
		})
	}
	if filter.ProviderID != "" {
		stmt = stmt.Where(goqu.Ex{
			"provider_id": filter.ProviderID,
		})
	}
	if filter.PlanID != "" {
		stmt = stmt.Where(goqu.Ex{
			"plan_id": filter.PlanID,
		})
	}
	if filter.State != "" {
		stmt = stmt.Where(goqu.Ex{
			"state": filter.State,
		})
	}
	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", parseErr, err)
	}

	var subscriptionModels []Subscription
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_SUBSCRIPTIONS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &subscriptionModels, query, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	var subscriptions []subscription.Subscription
	for _, subscriptionModel := range subscriptionModels {
		subscription, err := subscriptionModel.transform()
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, subscription)
	}
	return subscriptions, nil
}

func (r BillingSubscriptionRepository) Delete(ctx context.Context, id string) error {
	stmt := dialect.Delete(TABLE_BILLING_SUBSCRIPTIONS).Where(goqu.Ex{
		"id": id,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", parseErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_SUBSCRIPTIONS, "Delete", func(ctx context.Context) error {
		_, err := r.dbc.ExecContext(ctx, query, params...)
		return err
	}); err != nil {
		return fmt.Errorf("%w: %s", dbErr, err)
	}
	return nil
}

// createSubscriptionAuditRecord creates an audit record for subscription operations
func (r BillingSubscriptionRepository) createSubscriptionAuditRecord(
	ctx context.Context,
	tx *sqlx.Tx,
	event pkgAuditRecord.Event,
	sub subscriptionWithCustomer,
	targetMetadata map[string]interface{},
	occurredAt time.Time,
) error {
	auditRecord := BuildAuditRecord(
		ctx,
		event,
		AuditResource{
			ID:   sub.SubscriptionID,
			Type: pkgAuditRecord.BillingCustomerType,
			Name: sub.CustomerName,
		},
		&AuditTarget{
			ID:       sub.ID,
			Type:     pkgAuditRecord.BillingSubscriptionType,
			Metadata: targetMetadata,
		},
		sub.CustomerOrgID,
		nil,
		occurredAt,
	)

	return InsertAuditRecordInTx(ctx, tx, auditRecord)
}
