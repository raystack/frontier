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
	"github.com/jmoiron/sqlx/types"
	"github.com/raystack/frontier/billing/subscription"
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

type BillingSubscriptionRepository struct {
	dbc *db.Client
}

func NewBillingSubscriptionRepository(dbc *db.Client) *BillingSubscriptionRepository {
	return &BillingSubscriptionRepository{
		dbc: dbc,
	}
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
	query, params, err := dialect.Insert(TABLE_BILLING_SUBSCRIPTIONS).
		Rows(record).Returning(&Subscription{}).ToSQL()
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var subscriptionModel Subscription
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_SUBSCRIPTIONS, "Create", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&subscriptionModel)
	}); err != nil {
		return subscription.Subscription{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return subscriptionModel.transform()
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

	query, params, err := dialect.Update(TABLE_BILLING_SUBSCRIPTIONS).Set(updateRecord).Where(goqu.Ex{
		"id": toUpdate.ID,
	}).Returning(&Subscription{}).ToSQL()
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var customerModel Subscription
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_SUBSCRIPTIONS, "Update", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&customerModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return subscription.Subscription{}, subscription.ErrNotFound
		case errors.Is(err, ErrInvalidTextRepresentation):
			return subscription.Subscription{}, subscription.ErrInvalidUUID
		default:
			return subscription.Subscription{}, fmt.Errorf("%w: %w", txnErr, err)
		}
	}

	return customerModel.transform()
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
