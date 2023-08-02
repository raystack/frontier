package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/domain"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/log"
)

type DomainRepository struct {
	log log.Logger
	dbc *db.Client
	Now func() time.Time
}

func NewDomainRepository(logger log.Logger, dbc *db.Client) *DomainRepository {
	return &DomainRepository{
		dbc: dbc,
		log: logger,
		Now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *DomainRepository) Create(ctx context.Context, domain *domain.Domain) error {
	if domain.ID == "" {
		domain.ID = uuid.New().String()
	}
	if domain.CreatedAt.IsZero() {
		domain.CreatedAt = s.Now()
	}

	query, params, err := dialect.Insert(TABLE_DOMAINS).Rows(
		goqu.Record{
			"id":          domain.ID,
			"org_id":      domain.OrgID,
			"name":        domain.Name,
			"token":       domain.Token,
			"verified":    domain.Verified,
			"verified_at": domain.VerifiedAt,
			"created_at":  domain.CreatedAt,
		}).OnConflict(goqu.DoUpdate("id", goqu.Record{
		"org_id":      domain.OrgID,
		"name":        domain.Name,
		"token":       domain.Token,
		"verified":    domain.Verified,
		"verified_at": domain.VerifiedAt,
		"created_at":  domain.CreatedAt,
	})).Returning(&Domain{}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", parseErr, err)
	}

	var domainModel Domain
	if err = s.dbc.WithTimeout(ctx, TABLE_DOMAINS, "Create", func(ctx context.Context) error {
		return s.dbc.QueryRowxContext(ctx, query, params...).StructScan(&domainModel)
	}); err != nil {
		err = checkPostgresError(err)
		return fmt.Errorf("%w: %s", dbErr, err)
	}

	return nil
}

func (s *FlowRepository) List(ctx context.Context, flt domain.Filter) ([]domain.Domain, error) {
	stmt := dialect.Select(
		goqu.I("d.id"),
		goqu.I("d.org_id"),
		goqu.I("d.name"),
		goqu.I("d.token"),
		goqu.I("d.verified"),
		goqu.I("d.verified_at"),
		goqu.I("d.created_at"),
	)
	if flt.OrgID != "" && flt.Verified {
		stmt = stmt.Where(goqu.Ex{
			"org_id":   flt.OrgID,
			"verified": flt.Verified,
		})
	} else if flt.OrgID != "" {
		stmt = stmt.Where(goqu.Ex{
			"org_id": flt.OrgID,
		})
	} else if flt.Verified {
		stmt = stmt.Where(goqu.Ex{
			"verified": flt.Verified,
		})
	}

	query, params, err := stmt.From(TABLE_DOMAINS).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", parseErr, err)
	}

	var domains []Domain
	if err = s.dbc.WithTimeout(ctx, TABLE_DOMAINS, "List", func(ctx context.Context) error {
		return s.dbc.SelectContext(ctx, &domains, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	var result []domain.Domain
	for _, d := range domains {
		transformedDomain := d.transform()
		result = append(result, transformedDomain)
	}

	return result, nil
}

func (s *DomainRepository) Get(ctx context.Context, id string) (*domain.Domain, error) {
	query, params, err := dialect.Select(
		goqu.I("d.id"),
		goqu.I("d.org_id"),
		goqu.I("d.name"),
		goqu.I("d.token"),
		goqu.I("d.verified"),
		goqu.I("d.verified_at"),
		goqu.I("d.created_at"),
	).From(TABLE_DOMAINS).Where(goqu.Ex{
		"id": id,
	}).ToSQL()

	if err != nil {
		return nil, fmt.Errorf("%w: %s", queryErr, err)
	}

	var domainModel Domain
	if err = s.dbc.WithTimeout(ctx, TABLE_DOMAINS, "Get", func(ctx context.Context) error {
		return s.dbc.QueryRowxContext(ctx, query, params...).StructScan(&domainModel)
	}); err != nil {
		err = checkPostgresError(err)
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	domain := domainModel.transform()
	return &domain, nil
}

func (s *DomainRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, params, err := dialect.Delete(TABLE_DOMAINS).Where(goqu.Ex{
		"id": id,
	}).Returning(&Domain{}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	return s.dbc.WithTimeout(ctx, TABLE_DOMAINS, "Delete", func(ctx context.Context) error {
		result, err := s.dbc.ExecContext(ctx, query, params...)
		if err != nil {
			err = checkPostgresError(err)
			return fmt.Errorf("%w: %s", dbErr, err)
		}

		if count, _ := result.RowsAffected(); count > 0 {
			return nil
		}

		return domain.ErrNotExist
	})
}

func (s *DomainRepository) Update(ctx context.Context, id string, toUpdate *domain.Domain) error {
	query, params, err := dialect.Update(TABLE_DOMAINS).Set(
		goqu.Record{
			"token":      toUpdate.Token,
			"verified":   toUpdate.Verified,
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{
		"id": id,
	}).Returning(&Domain{}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	var domainModel Domain
	if err = s.dbc.WithTimeout(ctx, TABLE_DOMAINS, "Update", func(ctx context.Context) error {
		return s.dbc.QueryRowxContext(ctx, query, params...).StructScan(&domainModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return domain.ErrNotExist
		default:
			return fmt.Errorf("%w: %s", dbErr, err)
		}
	}

	return nil
}
