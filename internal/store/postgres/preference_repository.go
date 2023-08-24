package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/raystack/frontier/core/preference"
	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/raystack/frontier/pkg/db"
)

type PreferenceRepository struct {
	dbc *db.Client
}

func NewPreferenceRepository(dbc *db.Client) *PreferenceRepository {
	return &PreferenceRepository{
		dbc: dbc,
	}
}

func (s *PreferenceRepository) Set(ctx context.Context, pref preference.Preference) (preference.Preference, error) {
	if pref.ID == "" {
		pref.ID = uuid.New().String()
	}
	query, params, err := dialect.Insert(TABLE_PREFERENCES).Rows(
		goqu.Record{
			"id":            pref.ID,
			"name":          pref.Name,
			"value":         pref.Value,
			"resource_type": pref.ResourceType,
			"resource_id":   pref.ResourceID,
			"updated_at":    goqu.L("NOW()"),
		}).OnConflict(goqu.DoUpdate("resource_type, resource_id, name",
		goqu.Record{
			"value":      pref.Value,
			"updated_at": time.Now().UTC(),
		},
	)).Returning(&Preference{}).ToSQL()
	if err != nil {
		return preference.Preference{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var prefModel Preference
	if err = s.dbc.WithTimeout(ctx, TABLE_PREFERENCES, "Set", func(ctx context.Context) error {
		return s.dbc.QueryRowxContext(ctx, query, params...).StructScan(&prefModel)
	}); err != nil {
		err = checkPostgresError(err)
		return preference.Preference{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return prefModel.transformToPreference(), nil
}

func (s *PreferenceRepository) Get(ctx context.Context, id uuid.UUID) (preference.Preference, error) {
	var prefModel Preference
	query, params, err := dialect.From(TABLE_PREFERENCES).Where(
		goqu.Ex{
			"id": id,
		},
	).ToSQL()
	if err != nil {
		return preference.Preference{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = s.dbc.WithTimeout(ctx, TABLE_PREFERENCES, "Get", func(ctx context.Context) error {
		return s.dbc.QueryRowxContext(ctx, query, params...).StructScan(&prefModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return preference.Preference{}, preference.ErrNotFound
		}
		return preference.Preference{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return prefModel.transformToPreference(), nil
}

func (s *PreferenceRepository) List(ctx context.Context, flt preference.Filter) ([]preference.Preference, error) {
	var fetchedPreferences []Preference
	stmt := dialect.From(TABLE_PREFERENCES)
	if flt.OrgID != "" {
		stmt = stmt.Where(goqu.Ex{
			"resource_type": schema.OrganizationNamespace,
			"resource_id":   flt.OrgID,
		})
	} else if flt.UserID != "" {
		stmt = stmt.Where(goqu.Ex{
			"resource_type": schema.UserPrincipal,
			"resource_id":   flt.UserID,
		})
	} else if flt.ProjectID != "" {
		stmt = stmt.Where(goqu.Ex{
			"resource_type": schema.ProjectNamespace,
			"resource_id":   flt.ProjectID,
		})
	} else if flt.GroupID != "" {
		stmt = stmt.Where(goqu.Ex{
			"resource_type": schema.GroupNamespace,
			"resource_id":   flt.GroupID,
		})
	} else {
		return nil, preference.ErrInvalidFilter
	}

	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = s.dbc.WithTimeout(ctx, TABLE_PREFERENCES, "List", func(ctx context.Context) error {
		return s.dbc.SelectContext(ctx, &fetchedPreferences, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedPreferences []preference.Preference
	for _, fetchedPref := range fetchedPreferences {
		transformedPreferences = append(transformedPreferences, fetchedPref.transformToPreference())
	}

	return transformedPreferences, nil
}
