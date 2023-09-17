package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/doug-martin/goqu/v9"
	"github.com/raystack/frontier/core/namespace"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/pkg/db"
)

type PolicyRepository struct {
	dbc *db.Client
}

func NewPolicyRepository(dbc *db.Client) *PolicyRepository {
	return &PolicyRepository{
		dbc: dbc,
	}
}

func (r PolicyRepository) buildListQuery() *goqu.SelectDataset {
	return dialect.Select(
		"p.id",
		"p.resource_type",
		"p.resource_id",
		"p.principal_id",
		"p.principal_type",
		"p.role_id",
	).From(goqu.T(TABLE_POLICIES).As("p"))
}

func (r PolicyRepository) Get(ctx context.Context, id string) (policy.Policy, error) {
	if strings.TrimSpace(id) == "" {
		return policy.Policy{}, policy.ErrInvalidID
	}

	query, params, err := r.buildListQuery().
		Where(
			goqu.Ex{
				"p.id": id,
			},
		).ToSQL()
	if err != nil {
		return policy.Policy{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var policyModel Policy
	if err = r.dbc.WithTimeout(ctx, TABLE_POLICIES, "Get", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &policyModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return policy.Policy{}, policy.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
			return policy.Policy{}, policy.ErrInvalidUUID
		default:
			return policy.Policy{}, err
		}
	}

	transformedPolicy, err := policyModel.transformToPolicy()
	if err != nil {
		return policy.Policy{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedPolicy, nil
}

func (r PolicyRepository) List(ctx context.Context, flt policy.Filter) ([]policy.Policy, error) {
	var fetchedPolicies []Policy
	stmt := r.buildListQuery()
	if flt.OrgID != "" {
		stmt = stmt.Where(goqu.Ex{
			"resource_id":   flt.OrgID,
			"resource_type": schema.OrganizationNamespace,
		})
	}

	if flt.GroupID != "" {
		stmt = stmt.Where(goqu.Ex{
			"resource_id":   flt.GroupID,
			"resource_type": schema.GroupNamespace,
		})
	}
	if flt.ProjectID != "" {
		stmt = stmt.Where(goqu.Ex{
			"resource_id":   flt.ProjectID,
			"resource_type": schema.ProjectNamespace,
		})
	}
	if flt.PrincipalID != "" {
		stmt = stmt.Where(goqu.Ex{
			"principal_id": flt.PrincipalID,
		})
	}
	if flt.PrincipalType != "" {
		stmt = stmt.Where(goqu.Ex{
			"principal_type": flt.PrincipalType,
		})
	}
	if flt.RoleID != "" {
		stmt = stmt.Where(goqu.Ex{
			"role_id": flt.RoleID,
		})
	}

	query, params, err := stmt.ToSQL()
	if err != nil {
		return []policy.Policy{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_POLICIES, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedPolicies, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return []policy.Policy{}, nil
		default:
			return []policy.Policy{}, err
		}
	}

	var transformedPolicies []policy.Policy
	for _, p := range fetchedPolicies {
		transformedPolicy, err := p.transformToPolicy()
		if err != nil {
			return []policy.Policy{}, fmt.Errorf("%w: %s", parseErr, err)
		}
		transformedPolicies = append(transformedPolicies, transformedPolicy)
	}

	return transformedPolicies, nil
}

func (r PolicyRepository) Upsert(ctx context.Context, pol policy.Policy) (string, error) {
	marshaledMetadata, err := json.Marshal(pol.Metadata)
	if err != nil {
		return "", fmt.Errorf("%w: %s", parseErr, err)
	}

	query, params, err := dialect.Insert(TABLE_POLICIES).Rows(
		goqu.Record{
			"role_id":        pol.RoleID,
			"resource_type":  pol.ResourceType,
			"resource_id":    pol.ResourceID,
			"principal_id":   pol.PrincipalID,
			"principal_type": pol.PrincipalType,
			"metadata":       marshaledMetadata,
		}).OnConflict(goqu.DoUpdate("role_id, resource_id, resource_type, principal_id, principal_type", goqu.Record{
		"metadata": marshaledMetadata,
	})).Returning("id").ToSQL()
	if err != nil {
		return "", fmt.Errorf("%w: %s", queryErr, err)
	}

	var policyID string
	if err = r.dbc.WithTimeout(ctx, TABLE_POLICIES, "Upsert", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).Scan(&policyID)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, ErrForeignKeyViolation):
			return "", fmt.Errorf("%w: %s", policy.ErrInvalidDetail, err)
		default:
			return "", fmt.Errorf("%w: %s", dbErr, err)
		}
	}

	return policyID, nil
}

func (r PolicyRepository) Update(ctx context.Context, toUpdate policy.Policy) (string, error) {
	if strings.TrimSpace(toUpdate.ID) == "" {
		return "", policy.ErrInvalidID
	}
	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return "", fmt.Errorf("%w: %s", parseErr, err)
	}

	query, params, err := dialect.Update(TABLE_POLICIES).Set(
		goqu.Record{
			"metadata":   marshaledMetadata,
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{
		"id": toUpdate.ID,
	}).Returning("id").ToSQL()
	if err != nil {
		return "", fmt.Errorf("%w: %s", queryErr, err)
	}

	var policyID string
	if err = r.dbc.WithTimeout(ctx, TABLE_POLICIES, "Update", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).Scan(&policyID)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return "", policy.ErrNotExist
		case errors.Is(err, ErrDuplicateKey):
			return "", policy.ErrConflict
		case errors.Is(err, ErrInvalidTextRepresentation):
			return "", policy.ErrInvalidUUID
		case errors.Is(err, ErrForeignKeyViolation):
			return "", namespace.ErrNotExist
		default:
			return "", err
		}
	}

	return policyID, nil
}

func (r PolicyRepository) Delete(ctx context.Context, id string) error {
	query, params, err := dialect.Delete(TABLE_POLICIES).Where(
		goqu.Ex{
			"id": id,
		},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_POLICIES, "Delete", func(ctx context.Context) error {
		if _, err = r.dbc.DB.ExecContext(ctx, query, params...); err != nil {
			return err
		}
		return nil
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return policy.ErrNotExist
		default:
			return err
		}
	}
	return nil
}
