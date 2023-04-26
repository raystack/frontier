package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	newrelic "github.com/newrelic/go-agent"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/db"
	"github.com/odpf/shield/pkg/uuid"
)

type UserRepository struct {
	dbc *db.Client
}

type joinUserMetadata struct {
	ID        string         `db:"id"`
	Name      string         `db:"name"`
	Email     string         `db:"email"`
	Slug      sql.NullString `db:"slug"`
	Key       any            `db:"key"`
	Value     sql.NullString `db:"value"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

func NewUserRepository(dbc *db.Client) *UserRepository {
	return &UserRepository{
		dbc: dbc,
	}
}

var notDisabledUserExp = goqu.Or(
	goqu.Ex{
		"state": nil,
	},
	goqu.Ex{
		"state": goqu.Op{"neq": user.Disabled},
	},
)

func (r UserRepository) GetByID(ctx context.Context, id string) (user.User, error) {
	if strings.TrimSpace(id) == "" {
		return user.User{}, user.ErrInvalidID
	}

	var fetchedUser User
	userQuery, params, err := dialect.From(TABLE_USERS).
		Where(goqu.Ex{
			"id": id,
		}).Where(notDisabledUserExp).ToSQL()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_USERS,
				Operation:  "GetByID",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return r.dbc.GetContext(ctx, &fetchedUser, userQuery, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, errDuplicateKey):
			return user.User{}, user.ErrConflict
		case errors.Is(err, sql.ErrNoRows):
			return user.User{}, user.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
			return user.User{}, user.ErrInvalidUUID
		default:
			return user.User{}, err
		}
	}

	metadataQuery, params, err := dialect.From(TABLE_METADATA).Select("key", "value").
		Where(goqu.Ex{
			"user_id": fetchedUser.ID,
		}).ToSQL()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	data := make(map[string]interface{})

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_METADATA,
				Operation:  "GetByUserID",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		metadata, err := r.dbc.QueryContext(ctx, metadataQuery)
		if err != nil {
			return err
		}

		for {
			var key string
			var valuejson string
			if !metadata.Next() {
				break
			}
			err := metadata.Scan(&key, &valuejson)
			if err != nil {
				return err
			}
			var value any
			err = json.Unmarshal([]byte(valuejson), &value)
			if err != nil {
				return err
			}
			data[key] = value
		}

		return nil
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, errDuplicateKey):
			return user.User{}, user.ErrConflict
		default:
			return user.User{}, err
		}
	}

	transformedUser, err := fetchedUser.transformToUser()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}
	transformedUser.Metadata = data

	return transformedUser, nil
}

func (r UserRepository) GetBySlug(ctx context.Context, slug string) (user.User, error) {
	if strings.TrimSpace(slug) == "" {
		return user.User{}, user.ErrMissingSlug
	}

	var fetchedUser User
	query, params, err := dialect.From(TABLE_USERS).
		Where(goqu.Ex{
			"slug": slug,
		}).ToSQL()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_USERS,
				Operation:  "GetByUserSlug",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}
		return r.dbc.GetContext(ctx, &fetchedUser, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return user.User{}, user.ErrNotExist
		default:
			return user.User{}, err
		}
	}

	metadataQuery, params, err := dialect.From(TABLE_METADATA).Select("key", "value").
		Where(goqu.Ex{
			"user_id": fetchedUser.ID,
		}).ToSQL()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	data := make(map[string]interface{})

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_METADATA,
				Operation:  "GetByUserID",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		metadata, err := r.dbc.QueryContext(ctx, metadataQuery)
		if err != nil {
			return err
		}

		for {
			var key string
			var valuejson string
			if !metadata.Next() {
				break
			}
			err := metadata.Scan(&key, &valuejson)
			if err != nil {
				return err
			}
			var value any
			err = json.Unmarshal([]byte(valuejson), &value)
			if err != nil {
				return err
			}
			data[key] = value
		}

		return nil
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, errDuplicateKey):
			return user.User{}, user.ErrConflict
		default:
			return user.User{}, err
		}
	}

	transformedUser, err := fetchedUser.transformToUser()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}
	transformedUser.Metadata = data

	return transformedUser, nil
}

func (r UserRepository) Create(ctx context.Context, usr user.User) (user.User, error) {
	if strings.TrimSpace(usr.Email) == "" || strings.TrimSpace(usr.Slug) == "" {
		return user.User{}, user.ErrInvalidDetails
	}

	tx, err := r.dbc.BeginTx(ctx, nil)
	if err != nil {
		return user.User{}, err
	}

	insertRow := goqu.Record{
		"name":  usr.Name,
		"email": usr.Email,
		"slug":  usr.Slug,
	}
	if usr.State != "" {
		insertRow["state"] = usr.State
	}
	createQuery, params, err := dialect.Insert(TABLE_USERS).Rows(insertRow).Returning("created_at", "deleted_at", "email", "id", "name", "slug", "state", "updated_at").ToSQL()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var userModel User
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_USERS,
				Operation:  "Create",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return tx.QueryRowContext(ctx, createQuery, params...).
			Scan(&userModel.CreatedAt,
				&userModel.DeletedAt,
				&userModel.Email,
				&userModel.ID,
				&userModel.Name,
				&userModel.Slug,
				&userModel.State,
				&userModel.UpdatedAt,
			)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, errDuplicateKey):
			return user.User{}, user.ErrConflict
		default:
			tx.Rollback()
			return user.User{}, err
		}
	}

	transformedUser, err := userModel.transformToUser()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var rows []interface{}
	for k, v := range usr.Metadata {
		valuejson, err := json.Marshal(v)
		if err != nil {
			valuejson = []byte{}
		}

		rows = append(rows, goqu.Record{
			"user_id": transformedUser.ID,
			"key":     k,
			"value":   valuejson,
		})
	}
	metadataQuery, _, err := dialect.Insert(TABLE_METADATA).Rows(rows...).ToSQL()
	if err != nil {
		return user.User{}, err
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_METADATA,
				Operation:  "Create",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		_, err := tx.ExecContext(ctx, metadataQuery, params...)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, errDuplicateKey):
			return user.User{}, user.ErrConflict
		case errors.Is(err, errForeignKeyViolation):
			re := regexp.MustCompile(`\(([^)]+)\) `)
			match := re.FindStringSubmatch(err.Error())
			if len(match) > 1 {
				return user.User{}, fmt.Errorf("%w:%s", user.ErrKeyDoesNotExists, match[1])
			}
			return user.User{}, user.ErrKeyDoesNotExists

		default:
			tx.Rollback()
			return user.User{}, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return user.User{}, err
	}

	transformedUser.Metadata = usr.Metadata
	return transformedUser, nil
}

func (r UserRepository) List(ctx context.Context, flt user.Filter) ([]user.User, error) {
	var fetchedJoinUserMetadata []joinUserMetadata

	var defaultLimit int32 = 50
	var defaultPage int32 = 1
	if flt.Limit < 1 {
		flt.Limit = defaultLimit
	}
	if flt.Page < 1 {
		flt.Page = defaultPage
	}
	offset := (flt.Page - 1) * flt.Limit

	sqlStmt := dialect.From(TABLE_USERS).
		LeftOuterJoin(goqu.T(TABLE_METADATA), goqu.On(goqu.Ex{"users.id": goqu.I("metadata.user_id")})).
		Select("users.id", "name", "email", "slug", "key", "value", "users.created_at", "users.updated_at")

	if len(flt.Keyword) != 0 {
		sqlStmt = sqlStmt.Where(goqu.Or(
			goqu.C("name").ILike(fmt.Sprintf("%%%s%%", flt.Keyword)),
			goqu.C("email").ILike(fmt.Sprintf("%%%s%%", flt.Keyword)),
		))
	}
	if len(flt.State) != 0 {
		sqlStmt = sqlStmt.Where(goqu.Ex{
			"state": flt.State.String(),
		})
	} else {
		sqlStmt = sqlStmt.Where(notDisabledUserExp)
	}

	query, params, err := sqlStmt.Limit(uint(flt.Limit)).Offset(uint(offset)).ToSQL()
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: fmt.Sprintf("%s.%s", TABLE_USERS, TABLE_METADATA),
				Operation:  "List",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return r.dbc.SelectContext(ctx, &fetchedJoinUserMetadata, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []user.User{}, nil
		}
		return []user.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	groupedMetadataByUser := make(map[string]user.User)
	for _, u := range fetchedJoinUserMetadata {
		if _, ok := groupedMetadataByUser[u.ID]; !ok {
			groupedMetadataByUser[u.ID] = user.User{}
		}
		currentUser := groupedMetadataByUser[u.ID]
		currentUser.ID = u.ID
		currentUser.Email = u.Email
		currentUser.Name = u.Name
		currentUser.Slug = u.Slug.String
		currentUser.CreatedAt = u.CreatedAt
		currentUser.UpdatedAt = u.UpdatedAt

		if currentUser.Metadata == nil {
			currentUser.Metadata = make(map[string]any)
		}

		if u.Key != nil {
			var value any
			err := json.Unmarshal([]byte(u.Value.String), &value)
			if err != nil {
				continue
			}

			currentUser.Metadata[u.Key.(string)] = value
		}

		groupedMetadataByUser[u.ID] = currentUser
	}

	var transformedUsers []user.User
	for _, user := range groupedMetadataByUser {
		transformedUsers = append(transformedUsers, user)
	}

	return transformedUsers, nil
}

func (r UserRepository) GetByIDs(ctx context.Context, userIDs []string) ([]user.User, error) {
	var fetchedUsers []User

	query, params, err := dialect.From(TABLE_USERS).Select("id", "name", "email", "slug", "state").Where(
		goqu.Ex{
			"id": goqu.Op{"in": userIDs},
		}).Where(notDisabledUserExp).ToSQL()
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_USERS,
				Operation:  "GetByIDs",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return r.dbc.SelectContext(ctx, &fetchedUsers, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return []user.User{}, user.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
			return []user.User{}, user.ErrInvalidUUID
		default:
			return []user.User{}, err
		}
	}

	var transformedUsers []user.User
	for _, u := range fetchedUsers {
		transformedUser, err := u.transformToUser()
		if err != nil {
			return nil, err
		}
		transformedUsers = append(transformedUsers, transformedUser)
	}

	return transformedUsers, nil
}

func (r UserRepository) UpdateByEmail(ctx context.Context, usr user.User) (user.User, error) {
	userMetadata := make(map[string]any)

	if strings.TrimSpace(usr.Email) == "" {
		return user.User{}, user.ErrInvalidEmail
	}

	var transformedUser user.User

	err := r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		updateQuery, params, err := dialect.Update(TABLE_USERS).Set(
			goqu.Record{
				"name":       usr.Name,
				"slug":       usr.Slug,
				"updated_at": goqu.L("now()"),
			}).Where(
			goqu.Ex{
				"email": usr.Email,
			},
		).Returning("created_at", "deleted_at", "email", "id", "name", "state", "slug", "updated_at").ToSQL()
		if err != nil {
			return fmt.Errorf("%w: %s", queryErr, err)
		}

		var userModel User
		if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
			nrCtx := newrelic.FromContext(ctx)
			if nrCtx != nil {
				nr := newrelic.DatastoreSegment{
					Product:    newrelic.DatastorePostgres,
					Collection: TABLE_USERS,
					Operation:  "UpdateByEmail",
					StartTime:  nrCtx.StartSegmentNow(),
				}
				defer nr.End()
			}

			return tx.QueryRowContext(ctx, updateQuery, params...).
				Scan(&userModel.CreatedAt,
					&userModel.DeletedAt,
					&userModel.Email,
					&userModel.ID,
					&userModel.Name,
					&userModel.State,
					&userModel.Slug,
					&userModel.UpdatedAt,
				)
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return user.ErrNotExist
			}
			return fmt.Errorf("%s: %w", txnErr, err)
		}

		transformedUser, err = userModel.transformToUser()
		if err != nil {
			return fmt.Errorf("%s: %w", parseErr, err)
		}

		if usr.Metadata != nil {
			existingMetadataQuery, params, err := dialect.From(TABLE_METADATA).Select("key", "value").
				Where(goqu.Ex{
					"user_id": transformedUser.ID,
				}).ToSQL()
			if err != nil {
				return fmt.Errorf("%w: %s", queryErr, err)
			}

			if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
				nrCtx := newrelic.FromContext(ctx)
				if nrCtx != nil {
					nr := newrelic.DatastoreSegment{
						Product:    newrelic.DatastorePostgres,
						Collection: TABLE_METADATA,
						Operation:  "GetByUserID",
						StartTime:  nrCtx.StartSegmentNow(),
					}
					defer nr.End()
				}

				metadata, err := r.dbc.QueryContext(ctx, existingMetadataQuery)
				if err != nil {
					return err
				}

				for {
					var key string
					var valuejson string
					if !metadata.Next() {
						break
					}
					err := metadata.Scan(&key, &valuejson)
					if err != nil {
						return err
					}

					var value any
					err = json.Unmarshal([]byte(valuejson), &value)
					if err != nil {
						return err
					}

					userMetadata[key] = value
				}

				return nil
			}); err != nil {
				err = checkPostgresError(err)
				switch {
				case errors.Is(err, errDuplicateKey):
					return user.ErrConflict
				default:
					return err
				}
			}

			metadataDeleteQuery, params, err := dialect.Delete(TABLE_METADATA).
				Where(
					goqu.Ex{
						"user_id": transformedUser.ID,
					},
				).ToSQL()
			if err != nil {
				return nil
			}

			if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
				nrCtx := newrelic.FromContext(ctx)
				if nrCtx != nil {
					nr := newrelic.DatastoreSegment{
						Product:    newrelic.DatastorePostgres,
						Collection: TABLE_METADATA,
						Operation:  "DeleteByUserID",
						StartTime:  nrCtx.StartSegmentNow(),
					}
					defer nr.End()
				}

				_, err := tx.ExecContext(ctx, metadataDeleteQuery, params...)
				if err != nil {
					return err
				}
				return nil
			}); err != nil {
				err = checkPostgresError(err)
				switch {
				case errors.Is(err, errDuplicateKey):
					return user.ErrConflict
				default:
					return err
				}
			}

			for key, value := range usr.Metadata {
				userMetadata[key] = value
			}

			var rows []interface{}
			for k, v := range userMetadata {
				valuejson, err := json.Marshal(v)
				if err != nil {
					return err
				}
				rows = append(rows, goqu.Record{
					"user_id": transformedUser.ID,
					"key":     k,
					"value":   valuejson,
				})
			}
			metadataQuery, params, err := dialect.Insert(TABLE_METADATA).Rows(rows...).ToSQL()
			if err != nil {
				return err
			}

			if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
				nrCtx := newrelic.FromContext(ctx)
				if nrCtx != nil {
					nr := newrelic.DatastoreSegment{
						Product:    newrelic.DatastorePostgres,
						Collection: TABLE_METADATA,
						Operation:  "Create",
						StartTime:  nrCtx.StartSegmentNow(),
					}
					defer nr.End()
				}

				_, err := tx.ExecContext(ctx, metadataQuery, params...)
				if err != nil {
					return err
				}
				return nil
			}); err != nil {
				err = checkPostgresError(err)
				switch {
				case errors.Is(err, errDuplicateKey):
					return user.ErrConflict
				default:
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return user.User{}, err
	}

	transformedUser.Metadata = userMetadata

	return transformedUser, nil
}

func (r UserRepository) UpdateByID(ctx context.Context, usr user.User) (user.User, error) {
	if usr.ID == "" || !uuid.IsValid(usr.ID) {
		return user.User{}, user.ErrInvalidID
	}
	if strings.TrimSpace(usr.Email) == "" || strings.TrimSpace(usr.Slug) == "" {
		return user.User{}, user.ErrInvalidDetails
	}

	var transformedUser user.User

	err := r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		query, params, err := dialect.Update(TABLE_USERS).Set(
			goqu.Record{
				"name":       usr.Name,
				"slug":       usr.Slug,
				"updated_at": goqu.L("now()"),
			}).Where(
			goqu.Ex{
				"id": usr.ID,
			},
		).Returning("created_at", "deleted_at", "email", "id", "name", "state", "slug", "updated_at").ToSQL()
		if err != nil {
			return fmt.Errorf("%w: %s", queryErr, err)
		}

		var userModel User
		if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
			nrCtx := newrelic.FromContext(ctx)
			if nrCtx != nil {
				nr := newrelic.DatastoreSegment{
					Product:    newrelic.DatastorePostgres,
					Collection: TABLE_USERS,
					Operation:  "UpdateByID",
					StartTime:  nrCtx.StartSegmentNow(),
				}
				defer nr.End()
			}

			return tx.QueryRowContext(ctx, query, params...).Scan(&userModel.CreatedAt,
				&userModel.DeletedAt,
				&userModel.Email,
				&userModel.ID,
				&userModel.Name,
				&userModel.State,
				&userModel.Slug,
				&userModel.UpdatedAt,
			)
		}); err != nil {
			err = checkPostgresError(err)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return user.ErrNotExist
			case errors.Is(err, errDuplicateKey):
				return user.ErrConflict
			default:
				return err
			}
		}

		transformedUser, err = userModel.transformToUser()
		if err != nil {
			return fmt.Errorf("%s: %w", parseErr, err)
		}

		metadataDeleteQuery, params, err := dialect.Delete(TABLE_METADATA).
			Where(
				goqu.Ex{
					"user_id": transformedUser.ID,
				},
			).ToSQL()
		if err != nil {
			return nil
		}

		if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
			nrCtx := newrelic.FromContext(ctx)
			if nrCtx != nil {
				nr := newrelic.DatastoreSegment{
					Product:    newrelic.DatastorePostgres,
					Collection: TABLE_METADATA,
					Operation:  "DeleteByUserID",
					StartTime:  nrCtx.StartSegmentNow(),
				}
				defer nr.End()
			}

			_, err := tx.ExecContext(ctx, metadataDeleteQuery, params...)
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
			err = checkPostgresError(err)
			switch {
			case errors.Is(err, errDuplicateKey):
				return user.ErrConflict
			default:
				return err
			}
		}

		if len(usr.Metadata) > 0 {
			var rows []interface{}

			for k, v := range usr.Metadata {
				valuejson, err := json.Marshal(v)
				if err != nil {
					valuejson = []byte{}
				}

				rows = append(rows, goqu.Record{
					"user_id": transformedUser.ID,
					"key":     k,
					"value":   valuejson,
				})
			}
			metadataQuery, _, err := dialect.Insert(TABLE_METADATA).Rows(rows...).ToSQL()
			if err != nil {
				return err
			}

			if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
				nrCtx := newrelic.FromContext(ctx)
				if nrCtx != nil {
					nr := newrelic.DatastoreSegment{
						Product:    newrelic.DatastorePostgres,
						Collection: TABLE_METADATA,
						Operation:  "Create",
						StartTime:  nrCtx.StartSegmentNow(),
					}
					defer nr.End()
				}

				_, err := tx.ExecContext(ctx, metadataQuery, params...)
				if err != nil {
					return err
				}
				return nil
			}); err != nil {
				err = checkPostgresError(err)
				switch {
				case errors.Is(err, errDuplicateKey):
					return user.ErrConflict
				default:
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		return user.User{}, err
	}

	transformedUser.Metadata = usr.Metadata
	return transformedUser, nil
}

func (r UserRepository) UpdateBySlug(ctx context.Context, usr user.User) (user.User, error) {
	if usr.Slug == "" {
		return user.User{}, user.ErrMissingSlug
	}

	if strings.TrimSpace(usr.Email) == "" {
		return user.User{}, user.ErrInvalidDetails
	}
	var transformedUser user.User

	err := r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		query, params, err := dialect.Update(TABLE_USERS).Set(
			goqu.Record{
				"name":       usr.Name,
				"slug":       usr.Slug,
				"updated_at": goqu.L("now()"),
			}).Where(
			goqu.Ex{
				"slug": usr.ID,
			},
		).Returning("created_at", "deleted_at", "email", "id", "name", "slug", "updated_at").ToSQL()
		if err != nil {
			return fmt.Errorf("%w: %s", queryErr, err)
		}

		var userModel User
		if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
			nrCtx := newrelic.FromContext(ctx)
			if nrCtx != nil {
				nr := newrelic.DatastoreSegment{
					Product:    newrelic.DatastorePostgres,
					Collection: TABLE_USERS,
					Operation:  "UpdateBySlug",
					StartTime:  nrCtx.StartSegmentNow(),
				}
				defer nr.End()
			}

			return tx.QueryRowContext(ctx, query, params...).Scan(&userModel.CreatedAt,
				&userModel.DeletedAt,
				&userModel.Email,
				&userModel.ID,
				&userModel.Name,
				&userModel.Slug,
				&userModel.UpdatedAt)
		}); err != nil {
			err = checkPostgresError(err)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return user.ErrNotExist
			case errors.Is(err, errDuplicateKey):
				return user.ErrConflict
			default:
				return err
			}
		}

		transformedUser, err = userModel.transformToUser()
		if err != nil {
			return fmt.Errorf("%s: %w", parseErr, err)
		}

		metadataDeleteQuery, params, err := dialect.Delete(TABLE_METADATA).
			Where(
				goqu.Ex{
					"user_id": transformedUser.ID,
				},
			).ToSQL()
		if err != nil {
			return nil
		}

		if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
			nrCtx := newrelic.FromContext(ctx)
			if nrCtx != nil {
				nr := newrelic.DatastoreSegment{
					Product:    newrelic.DatastorePostgres,
					Collection: TABLE_METADATA,
					Operation:  "DeleteByUserSlug",
					StartTime:  nrCtx.StartSegmentNow(),
				}
				defer nr.End()
			}

			_, err := tx.ExecContext(ctx, metadataDeleteQuery, params...)
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
			err = checkPostgresError(err)
			switch {
			case errors.Is(err, errDuplicateKey):
				return user.ErrConflict
			default:
				return err
			}
		}

		if len(usr.Metadata) > 0 {
			var rows []interface{}

			for k, v := range usr.Metadata {
				valuejson, err := json.Marshal(v)
				if err != nil {
					valuejson = []byte{}
				}

				rows = append(rows, goqu.Record{
					"user_id": transformedUser.ID,
					"key":     k,
					"value":   valuejson,
				})
			}
			metadataQuery, _, err := dialect.Insert(TABLE_METADATA).Rows(rows...).ToSQL()
			if err != nil {
				return err
			}

			if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
				nrCtx := newrelic.FromContext(ctx)
				if nrCtx != nil {
					nr := newrelic.DatastoreSegment{
						Product:    newrelic.DatastorePostgres,
						Collection: TABLE_METADATA,
						Operation:  "Create",
						StartTime:  nrCtx.StartSegmentNow(),
					}
					defer nr.End()
				}

				_, err := tx.ExecContext(ctx, metadataQuery, params...)
				if err != nil {
					return err
				}
				return nil
			}); err != nil {
				err = checkPostgresError(err)
				switch {
				case errors.Is(err, errDuplicateKey):
					return user.ErrConflict
				default:
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		return user.User{}, err
	}

	transformedUser.Metadata = usr.Metadata
	return transformedUser, nil
}

func (r UserRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	if strings.TrimSpace(email) == "" {
		return user.User{}, user.ErrInvalidEmail
	}

	var fetchedUser User
	data := make(map[string]any)

	query, params, err := dialect.From(TABLE_USERS).Where(
		goqu.Ex{
			"email": email,
		}).Where(notDisabledUserExp).ToSQL()

	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_USERS,
				Operation:  "GetByEmail",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return r.dbc.GetContext(ctx, &fetchedUser, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user.User{}, user.ErrNotExist
		}
		return user.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	metadataQuery, params, err := dialect.From(TABLE_METADATA).Select("key", "value").
		Where(goqu.Ex{
			"user_id": fetchedUser.ID,
		}).ToSQL()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_METADATA,
				Operation:  "GetByUserID",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		metadata, err := r.dbc.QueryContext(ctx, metadataQuery)
		if err != nil {
			return err
		}

		for {
			var key string
			var value any
			if !metadata.Next() {
				break
			}
			metadata.Scan(&key, &value)
			data[key] = value
		}

		return nil
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, errDuplicateKey):
			return user.User{}, user.ErrConflict
		default:
			return user.User{}, err
		}
	}

	transformedUser, err := fetchedUser.transformToUser()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	transformedUser.Metadata = data

	return transformedUser, nil
}

func (r UserRepository) SetState(ctx context.Context, id string, state user.State) error {
	query, params, err := dialect.Update(TABLE_USERS).Set(
		goqu.Record{
			"state": state.String(),
		}).Where(
		goqu.Ex{
			"id": id,
		},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_USERS,
				Operation:  "SetState",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		if _, err = r.dbc.DB.ExecContext(ctx, query, params...); err != nil {
			return err
		}
		return nil
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return user.ErrNotExist
		default:
			return err
		}
	}
	return nil
}

func (r UserRepository) Delete(ctx context.Context, id string) error {
	query, params, err := dialect.Delete(TABLE_USERS).Where(
		goqu.Ex{
			"id": id,
		},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_USERS,
				Operation:  "Delete",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		if _, err = r.dbc.DB.ExecContext(ctx, query, params...); err != nil {
			return err
		}
		return nil
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return user.ErrNotExist
		default:
			return err
		}
	}
	return nil
}
