package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/raystack/frontier/core/namespace"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/auditrecord"
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
		return policy.Policy{}, fmt.Errorf("%w: %w", queryErr, err)
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
		return policy.Policy{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	return transformedPolicy, nil
}

func applyListFilter(stmt *goqu.SelectDataset, flt policy.Filter) *goqu.SelectDataset {
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
	} else if len(flt.PrincipalIDs) > 0 {
		stmt = stmt.Where(goqu.Ex{
			"principal_id": flt.PrincipalIDs,
		})
	}
	if flt.PrincipalType != "" {
		stmt = stmt.Where(goqu.Ex{
			"principal_type": flt.PrincipalType,
		})
	}
	if flt.ResourceType != "" {
		stmt = stmt.Where(goqu.Ex{
			"resource_type": flt.ResourceType,
		})
	}
	if flt.RoleID != "" {
		stmt = stmt.Where(goqu.Ex{
			"role_id": flt.RoleID,
		})
	}
	if len(flt.RoleIDs) > 0 {
		stmt = stmt.Where(goqu.Ex{
			"role_id": flt.RoleIDs,
		})
	}
	return stmt
}

func (r PolicyRepository) List(ctx context.Context, flt policy.Filter) ([]policy.Policy, error) {
	var fetchedPolicies []Policy
	stmt := r.buildListQuery()
	stmt = applyListFilter(stmt, flt)

	query, params, err := stmt.ToSQL()
	if err != nil {
		return []policy.Policy{}, fmt.Errorf("%w: %w", queryErr, err)
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
			return []policy.Policy{}, fmt.Errorf("%w: %w", parseErr, err)
		}
		transformedPolicies = append(transformedPolicies, transformedPolicy)
	}

	return transformedPolicies, nil
}

func (r PolicyRepository) Count(ctx context.Context, flt policy.Filter) (int64, error) {
	var count int64
	stmt := dialect.Select(goqu.COUNT(goqu.Star()).As("count")).From(goqu.T(TABLE_POLICIES).As("p"))
	stmt = applyListFilter(stmt, flt)

	query, params, err := stmt.ToSQL()
	if err != nil {
		return count, fmt.Errorf("%w: %w", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_POLICIES, "Count", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &count, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return count, nil
		default:
			return count, err
		}
	}
	return count, nil
}

func (r PolicyRepository) Upsert(ctx context.Context, pol policy.Policy) (policy.Policy, error) {
	marshaledMetadata, err := json.Marshal(pol.Metadata)
	if err != nil {
		return policy.Policy{}, fmt.Errorf("%w: %w", parseErr, err)
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
		"metadata":   marshaledMetadata,
		"updated_at": goqu.L("now()"),
	})).Returning(&PolicyCols{}).ToSQL()
	if err != nil {
		return policy.Policy{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	// Check if policy exists before upsert
	_, exists := r.getPolicyByConstraint(ctx, pol)

	var policyDB Policy
	if err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_POLICIES, "Upsert", func(ctx context.Context) error {
			if err := tx.QueryRowxContext(ctx, query, params...).StructScan(&policyDB); err != nil {
				return err
			}

			var (
				event              auditrecord.Event
				timestamp          time.Time
				additionalMetadata map[string]any
			)
			if exists {
				event = auditrecord.PolicyUpdatedEvent
				timestamp = policyDB.UpdatedAt
				additionalMetadata = map[string]any{
					"updated_metadata": pol.Metadata,
				}
			} else {
				event = auditrecord.PolicyCreatedEvent
				timestamp = policyDB.CreatedAt
			}

			auditRecord := r.buildPolicyAuditRecord(ctx, tx, event, policyDB, timestamp, additionalMetadata)
			return InsertAuditRecordInTx(ctx, tx, auditRecord)
		})
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, ErrForeignKeyViolation):
			return policy.Policy{}, fmt.Errorf("%w: %w", policy.ErrInvalidDetail, err)
		default:
			return policy.Policy{}, fmt.Errorf("%w: %s", dbErr, err)
		}
	}

	return policyDB.transformToPolicy()
}

func (r PolicyRepository) Update(ctx context.Context, toUpdate policy.Policy) (string, error) {
	if strings.TrimSpace(toUpdate.ID) == "" {
		return "", policy.ErrInvalidID
	}

	// Fetch existing policy for audit record
	existingPolicy, err := r.Get(ctx, toUpdate.ID)
	if err != nil {
		return "", err
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
	}).Returning("id", "updated_at").ToSQL()
	if err != nil {
		return "", fmt.Errorf("%w: %s", queryErr, err)
	}

	var policyID string
	var updatedAt time.Time
	if err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_POLICIES, "Update", func(ctx context.Context) error {
			if err := tx.QueryRowxContext(ctx, query, params...).Scan(&policyID, &updatedAt); err != nil {
				return err
			}

			policyDB := Policy{
				ID:            existingPolicy.ID,
				RoleID:        existingPolicy.RoleID,
				ResourceID:    existingPolicy.ResourceID,
				ResourceType:  existingPolicy.ResourceType,
				PrincipalID:   existingPolicy.PrincipalID,
				PrincipalType: existingPolicy.PrincipalType,
			}
			auditRecord := r.buildPolicyAuditRecord(ctx, tx, auditrecord.PolicyUpdatedEvent, policyDB, updatedAt, map[string]any{
				"updated_metadata": toUpdate.Metadata,
			})
			return InsertAuditRecordInTx(ctx, tx, auditRecord)
		})
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
	// Fetch policy for audit record
	existingPolicy, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	if err := r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_POLICIES, "Delete", func(ctx context.Context) error {
			deleteQuery, deleteParams, err := dialect.Delete(TABLE_POLICIES).
				Where(goqu.Ex{"id": id}).
				ToSQL()
			if err != nil {
				return fmt.Errorf("%w: %w", queryErr, err)
			}
			if _, err := tx.ExecContext(ctx, deleteQuery, deleteParams...); err != nil {
				return err
			}

			policyDB := Policy{
				ID:            existingPolicy.ID,
				RoleID:        existingPolicy.RoleID,
				ResourceID:    existingPolicy.ResourceID,
				ResourceType:  existingPolicy.ResourceType,
				PrincipalID:   existingPolicy.PrincipalID,
				PrincipalType: existingPolicy.PrincipalType,
			}
			auditRecord := r.buildPolicyAuditRecord(ctx, tx, auditrecord.PolicyDeletedEvent, policyDB, time.Now(), nil)
			return InsertAuditRecordInTx(ctx, tx, auditRecord)
		})
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

func (r PolicyRepository) GroupMemberCount(ctx context.Context, groupIDs []string) ([]policy.MemberCount, error) {
	if len(groupIDs) == 0 {
		return nil, policy.ErrInvalidID
	}
	stmt := goqu.From("policies").
		Select(goqu.I("resource_id").As("id"), goqu.COUNT(goqu.DISTINCT(goqu.I("principal_id"))).As("count")).
		Where(goqu.Ex{
			"resource_type": schema.GroupNamespace,
			"resource_id":   groupIDs,
			"principal_type": []string{
				schema.UserPrincipal,
				schema.ServiceUserPrincipal,
			},
		}).
		GroupBy("resource_id")

	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", queryErr, err)
	}

	var result []policy.MemberCount
	if err = r.dbc.WithTimeout(ctx, TABLE_POLICIES, "GroupMemberCount", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &result, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return result, nil
		default:
			return nil, err
		}
	}

	return result, nil
}

func (r PolicyRepository) ProjectMemberCount(ctx context.Context, projectIDs []string) ([]policy.MemberCount, error) {
	if len(projectIDs) == 0 {
		return nil, policy.ErrInvalidID
	}
	stmt := goqu.From("policies").
		Select(goqu.I("resource_id").As("id"), goqu.COUNT(goqu.DISTINCT(goqu.I("principal_id"))).As("count")).
		Where(goqu.Ex{
			"resource_type": schema.ProjectNamespace,
			"resource_id":   projectIDs,
			"principal_type": []string{
				schema.UserPrincipal, // check only for human users and not service users or groups
			},
		}).
		GroupBy("resource_id")

	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", queryErr, err)
	}

	var result []policy.MemberCount
	if err = r.dbc.WithTimeout(ctx, TABLE_POLICIES, "ProjectMemberCount", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &result, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return result, nil
		default:
			return nil, err
		}
	}

	return result, nil
}

func (r PolicyRepository) OrgMemberCount(ctx context.Context, id string) (policy.MemberCount, error) {
	if len(id) == 0 {
		return policy.MemberCount{}, policy.ErrInvalidID
	}
	stmt := goqu.From("policies").
		Select(goqu.I("resource_id").As("id"), goqu.COUNT(goqu.DISTINCT(goqu.I("principal_id"))).As("count")).
		Where(goqu.Ex{
			"resource_type": schema.OrganizationNamespace,
			"resource_id":   id,
			"principal_type": []string{
				schema.UserPrincipal,
			},
		}).
		GroupBy("resource_id")

	query, params, err := stmt.ToSQL()
	if err != nil {
		return policy.MemberCount{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var result policy.MemberCount
	if err = r.dbc.WithTimeout(ctx, TABLE_POLICIES, "OrgMemberCount", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &result, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return result, nil
		default:
			return result, err
		}
	}

	return result, nil
}

// buildPolicyAuditRecord builds an audit record for policy events
func (r PolicyRepository) buildPolicyAuditRecord(ctx context.Context, tx *sqlx.Tx, event auditrecord.Event, pol Policy, timestamp time.Time, additionalMetadata map[string]any) AuditRecord {
	orgID, resourceName := r.getResourceInfo(ctx, tx, pol.ResourceType, pol.ResourceID)

	targetMetadata := map[string]any{
		"role_id":        pol.RoleID,
		"principal_id":   pol.PrincipalID,
		"principal_type": pol.PrincipalType,
	}
	for k, v := range additionalMetadata {
		targetMetadata[k] = v
	}

	return BuildAuditRecord(
		ctx,
		event,
		AuditResource{
			ID:   pol.ResourceID,
			Type: mapResourceTypeToAuditType(pol.ResourceType),
			Name: resourceName,
		},
		&AuditTarget{
			ID:       pol.ID,
			Type:     auditrecord.PolicyType,
			Metadata: targetMetadata,
		},
		orgID,
		nil,
		timestamp,
	)
}

// getPolicyByConstraint fetches a policy by unique constraint fields
// Returns the policy and true if found, empty policy and false if not found
func (r PolicyRepository) getPolicyByConstraint(ctx context.Context, pol policy.Policy) (Policy, bool) {
	query, params, _ := dialect.From(TABLE_POLICIES).
		Select("id", "resource_type", "resource_id", "principal_id", "principal_type", "role_id").
		Where(goqu.Ex{
			"role_id":        pol.RoleID,
			"resource_id":    pol.ResourceID,
			"resource_type":  pol.ResourceType,
			"principal_id":   pol.PrincipalID,
			"principal_type": pol.PrincipalType,
		}).
		Limit(1).
		ToSQL()

	var existing Policy
	if err := r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&existing); err != nil {
		return Policy{}, false
	}
	return existing, true
}

// getResourceInfo fetches org ID and resource name based on resource type
func (r PolicyRepository) getResourceInfo(ctx context.Context, tx *sqlx.Tx, resourceType, resourceID string) (string, string) {
	var orgID, resourceName string
	switch resourceType {
	case schema.OrganizationNamespace:
		orgID = resourceID
		orgQuery, orgParams, _ := dialect.From(TABLE_ORGANIZATIONS).
			Select("title").
			Where(goqu.Ex{"id": resourceID}).
			ToSQL()
		_ = tx.QueryRowContext(ctx, orgQuery, orgParams...).Scan(&resourceName)
	case schema.ProjectNamespace:
		projQuery, projParams, _ := dialect.From(TABLE_PROJECTS).
			Select("org_id", "title").
			Where(goqu.Ex{"id": resourceID}).
			ToSQL()
		_ = tx.QueryRowContext(ctx, projQuery, projParams...).Scan(&orgID, &resourceName)
	case schema.GroupNamespace:
		grpQuery, grpParams, _ := dialect.From(TABLE_GROUPS).
			Select("org_id", "title").
			Where(goqu.Ex{"id": resourceID}).
			ToSQL()
		_ = tx.QueryRowContext(ctx, grpQuery, grpParams...).Scan(&orgID, &resourceName)
	}
	return orgID, resourceName
}

// mapResourceTypeToAuditType maps resource namespace to audit entity type
func mapResourceTypeToAuditType(resourceType string) auditrecord.EntityType {
	switch resourceType {
	case schema.OrganizationNamespace:
		return auditrecord.OrganizationType
	case schema.ProjectNamespace:
		return auditrecord.ProjectType
	case schema.GroupNamespace:
		return auditrecord.GroupType
	default:
		return auditrecord.EntityType(resourceType)
	}
}
