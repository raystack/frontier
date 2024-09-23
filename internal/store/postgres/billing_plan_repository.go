package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
	"github.com/raystack/frontier/billing/plan"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/pkg/db"
)

type Plan struct {
	ID string `db:"id"`

	Name           string  `db:"name"`
	Title          *string `db:"title"`
	Description    *string `db:"description"`
	Interval       *string `db:"interval"`
	OnStartCredits int64   `db:"on_start_credits"`

	State     string             `db:"state"`
	TrialDays *int64             `db:"trial_days"`
	Metadata  types.NullJSONText `db:"metadata"`

	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

type PlanProductRow struct {
	PlanID string `db:"plan_id"`

	PlanName           string  `db:"plan_name"`
	PlanTitle          *string `db:"plan_title"`
	PlanDescription    *string `db:"plan_description"`
	PlanInterval       *string `db:"plan_interval"`
	PlanOnStartCredits int64   `db:"plan_on_start_credits"`

	PlanState     string             `db:"plan_state"`
	PlanTrialDays *int64             `db:"plan_trial_days"`
	PlanMetadata  types.NullJSONText `db:"plan_metadata"`

	PlanCreatedAt time.Time  `db:"plan_created_at"`
	PlanUpdatedAt time.Time  `db:"plan_updated_at"`
	PlanDeletedAt *time.Time `db:"plan_deleted_at"`

	ProductID          string         `db:"product_id"`
	ProductProviderID  string         `db:"product_provider_id"`
	ProductPlanIDs     pq.StringArray `db:"product_plan_ids"`
	ProductName        string         `db:"product_name"`
	ProductTitle       *string        `db:"product_title"`
	ProductDescription *string        `db:"product_description"`

	ProductBehavior string             `db:"product_behavior"`
	ProductConfig   BehaviorConfig     `db:"product_config"`
	ProductState    string             `db:"product_state"`
	ProductMetadata types.NullJSONText `db:"product_metadata"`

	ProductCreatedAt time.Time  `db:"product_created_at"`
	ProductUpdatedAt time.Time  `db:"product_updated_at"`
	ProductDeletedAt *time.Time `db:"product_deleted_at"`
}

func (pr PlanProductRow) getPlan() (plan.Plan, error) {
	pln := Plan{
		ID:             pr.PlanID,
		Name:           pr.PlanName,
		Title:          pr.PlanTitle,
		Description:    pr.PlanDescription,
		Interval:       pr.PlanInterval,
		OnStartCredits: pr.PlanOnStartCredits,
		State:          pr.PlanState,
		TrialDays:      pr.PlanTrialDays,
		Metadata:       pr.PlanMetadata,

		CreatedAt: pr.PlanCreatedAt,
		UpdatedAt: pr.PlanUpdatedAt,
		DeletedAt: pr.PlanDeletedAt,
	}

	return pln.transform()
}

func (pr PlanProductRow) getProduct() (product.Product, error) {
	prod := Product{
		ID:          pr.ProductID,
		ProviderID:  pr.ProductProviderID,
		PlanIDs:     pr.ProductPlanIDs,
		Name:        pr.ProductName,
		Title:       pr.ProductTitle,
		Description: pr.ProductDescription,
		Behavior:    pr.ProductBehavior,
		Config:      pr.ProductConfig,
		State:       pr.ProductState,
		Metadata:    pr.ProductMetadata,
		CreatedAt:   pr.ProductCreatedAt,
		UpdatedAt:   pr.ProductUpdatedAt,
		DeletedAt:   pr.ProductDeletedAt,
	}

	return prod.transform()
}

func (c Plan) transform() (plan.Plan, error) {
	var unmarshalledMetadata map[string]any
	if c.Metadata.Valid {
		if err := c.Metadata.Unmarshal(&unmarshalledMetadata); err != nil {
			return plan.Plan{}, err
		}
	}
	var planTitle string
	if c.Title != nil {
		planTitle = *c.Title
	}
	var planDescription string
	if c.Description != nil {
		planDescription = *c.Description
	}
	var recurringInterval string
	if c.Interval != nil {
		recurringInterval = *c.Interval
	}
	var trialDays int64
	if c.TrialDays != nil {
		trialDays = *c.TrialDays
	}
	return plan.Plan{
		ID:             c.ID,
		Name:           c.Name,
		Title:          planTitle,
		Description:    planDescription,
		Interval:       recurringInterval,
		OnStartCredits: c.OnStartCredits,
		State:          c.State,
		TrialDays:      trialDays,
		Metadata:       unmarshalledMetadata,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
		DeletedAt:      c.DeletedAt,
	}, nil
}

type BillingPlanRepository struct {
	dbc *db.Client
}

func NewBillingPlanRepository(dbc *db.Client) *BillingPlanRepository {
	return &BillingPlanRepository{
		dbc: dbc,
	}
}

func (r BillingPlanRepository) Create(ctx context.Context, toCreate plan.Plan) (plan.Plan, error) {
	if toCreate.Metadata == nil {
		toCreate.Metadata = make(map[string]any)
	}
	marshaledMetadata, err := json.Marshal(toCreate.Metadata)
	if err != nil {
		return plan.Plan{}, err
	}
	if toCreate.ID == "" {
		toCreate.ID = uuid.New().String()
	}

	query, params, err := dialect.Insert(TABLE_BILLING_PLANS).Rows(
		goqu.Record{
			"id":               toCreate.ID,
			"name":             toCreate.Name,
			"title":            toCreate.Title,
			"description":      toCreate.Description,
			"interval":         toCreate.Interval,
			"on_start_credits": toCreate.OnStartCredits,
			"trial_days":       toCreate.TrialDays,
			"state":            toCreate.State,
			"metadata":         marshaledMetadata,
			"updated_at":       goqu.L("now()"),
		}).Returning(&Plan{}).ToSQL()
	if err != nil {
		return plan.Plan{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var planModel Plan
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_PLANS, "Create", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&planModel)
	}); err != nil {
		return plan.Plan{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return planModel.transform()
}

func (r BillingPlanRepository) GetByID(ctx context.Context, id string) (plan.Plan, error) {
	stmt := dialect.Select().From(TABLE_BILLING_PLANS).Where(goqu.Ex{
		"id": id,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return plan.Plan{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var planModel Plan
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_PLANS, "GetByID", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&planModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return plan.Plan{}, plan.ErrNotFound
		}
		return plan.Plan{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return planModel.transform()
}

func (r BillingPlanRepository) GetByName(ctx context.Context, name string) (plan.Plan, error) {
	stmt := dialect.Select().From(TABLE_BILLING_PLANS).Where(goqu.Ex{
		"name": name,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return plan.Plan{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var planModel Plan
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_PLANS, "GetByName", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&planModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return plan.Plan{}, plan.ErrNotFound
		}
		return plan.Plan{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return planModel.transform()
}

func (r BillingPlanRepository) UpdateByName(ctx context.Context, toUpdate plan.Plan) (plan.Plan, error) {
	if strings.TrimSpace(toUpdate.Name) == "" {
		return plan.Plan{}, plan.ErrInvalidName
	}
	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return plan.Plan{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	if toUpdate.State == "" {
		toUpdate.State = "active"
	}

	query, params, err := dialect.Update(TABLE_BILLING_PLANS).Set(
		goqu.Record{
			"title":            toUpdate.Title,
			"description":      toUpdate.Description,
			"on_start_credits": toUpdate.OnStartCredits,
			"trial_days":       toUpdate.TrialDays,
			"metadata":         marshaledMetadata,
			"state":            toUpdate.State,
			"updated_at":       goqu.L("now()"),
		}).Where(goqu.Ex{
		"name": toUpdate.Name,
	}).Returning(&Plan{}).ToSQL()
	if err != nil {
		return plan.Plan{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var planModel Plan
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_PLANS, "UpdateByName", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&planModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return plan.Plan{}, plan.ErrNotFound
		case errors.Is(err, ErrInvalidTextRepresentation):
			return plan.Plan{}, plan.ErrInvalidUUID
		default:
			return plan.Plan{}, fmt.Errorf("%s: %w", txnErr, err)
		}
	}

	return planModel.transform()
}

func (r BillingPlanRepository) List(ctx context.Context, filter plan.Filter) ([]plan.Plan, error) {
	stmt := dialect.Select().From(TABLE_BILLING_PLANS)
	var ids []string
	var names []string
	if len(filter.IDs) > 0 {
		if _, err := uuid.Parse(filter.IDs[0]); err == nil {
			ids = filter.IDs
		} else {
			names = filter.IDs
		}
	}
	if len(ids) > 0 {
		stmt = stmt.Where(goqu.Ex{
			"id": ids,
		})
	}
	if len(names) > 0 {
		stmt = stmt.Where(goqu.Ex{
			"name": names,
		})
	}
	if filter.Interval != "" {
		stmt = stmt.Where(goqu.Ex{
			"interval": filter.Interval,
		})
	}
	if filter.State == "" {
		filter.State = "active"
	}
	stmt = stmt.Where(goqu.Ex{
		"state": filter.State,
	})

	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", parseErr, err)
	}

	var planModels []Plan
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_PLANS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &planModels, query, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	plans := make([]plan.Plan, 0, len(planModels))
	for _, planModel := range planModels {
		plan, err := planModel.transform()
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

func (r BillingPlanRepository) ListWithProducts(ctx context.Context, filter plan.Filter) ([]plan.Plan, error) {
	pln := goqu.T(TABLE_BILLING_PLANS).As("plan")
	prd := goqu.T(TABLE_BILLING_PRODUCTS).As("product")
	stmt := dialect.From(pln).
		Join(
			prd,
			goqu.On(
				goqu.L("CAST(plan.id AS text)").Eq(goqu.L("ANY(product.plan_ids)")),
			),
		).Select(
		pln.Col("id").As("plan_id"),
		pln.Col("name").As("plan_name"),
		pln.Col("title").As("plan_title"),
		pln.Col("description").As("plan_description"),
		pln.Col("interval").As("plan_interval"),
		pln.Col("on_start_credits").As("plan_on_start_credits"),
		pln.Col("state").As("plan_state"),
		pln.Col("trial_days").As("plan_trial_days"),
		pln.Col("metadata").As("plan_metadata"),
		pln.Col("created_at").As("plan_created_at"),
		pln.Col("updated_at").As("plan_updated_at"),
		prd.Col("deleted_at").As("plan_deleted_at"),
		prd.Col("id").As("product_id"),
		prd.Col("provider_id").As("product_provider_id"),
		prd.Col("name").As("product_name"),
		prd.Col("title").As("product_title"),
		prd.Col("description").As("product_description"),
		prd.Col("title").As("product_behavior"),
		prd.Col("config").As("product_config"),
		prd.Col("state").As("product_state"),
		prd.Col("metadata").As("product_metadata"),
		prd.Col("created_at").As("product_created_at"),
		prd.Col("updated_at").As("product_updated_at"),
		prd.Col("deleted_at").As("product_deleted_at"),
	)

	var ids []string
	var names []string
	if len(filter.IDs) > 0 {
		if _, err := uuid.Parse(filter.IDs[0]); err == nil {
			ids = filter.IDs
		} else {
			names = filter.IDs
		}
	}
	if len(ids) > 0 {
		stmt = stmt.Where(goqu.Ex{
			"plan.id": ids,
		})
	}
	if len(names) > 0 {
		stmt = stmt.Where(goqu.Ex{
			"plan.name": names,
		})
	}
	if filter.Interval != "" {
		stmt = stmt.Where(goqu.Ex{
			"plan.interval": filter.Interval,
		})
	}
	if filter.State == "" {
		filter.State = "active"
	}
	stmt = stmt.Where(goqu.Ex{
		"plan.state": filter.State,
	})

	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", parseErr, err)
	}

	var detailedPlans []PlanProductRow
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_PLANS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &detailedPlans, query, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	planMap := map[string]plan.Plan{}

	for _, dbResult := range detailedPlans {
		pln, err := dbResult.getPlan()
		if err != nil {
			return nil, err
		}

		prod, err := dbResult.getProduct()
		if err != nil {
			return nil, err
		}

		planToReturn, exists := planMap[pln.ID]
		if exists {
			planToReturn.Products = append(planToReturn.Products, prod)
		} else {
			pln.Products = append(pln.Products, prod)
			planMap[pln.ID] = pln
		}
	}

	toReturn := []plan.Plan{}
	for _, item := range planMap {
		toReturn = append(toReturn, item)
	}

	return toReturn, nil
}
