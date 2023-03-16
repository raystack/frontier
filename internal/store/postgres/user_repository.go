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
	"github.com/goto/shield/core/user"
	"github.com/goto/shield/pkg/db"
	"github.com/goto/shield/pkg/uuid"
	"github.com/jmoiron/sqlx"
	newrelic "github.com/newrelic/go-agent"
)

type UserRepository struct {
	dbc *db.Client
}

type joinUserMetadata struct {
	ID        string         `db:"id"`
	Name      string         `db:"name"`
	Email     string         `db:"email"`
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

func (r UserRepository) GetByID(ctx context.Context, id string) (user.User, error) {
	if strings.TrimSpace(id) == "" {
		return user.User{}, user.ErrInvalidID
	}

	var fetchedUser User
	userQuery, params, err := dialect.From(TABLE_USERS).
		Where(goqu.Ex{
			"id": id,
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

func (r UserRepository) Create(ctx context.Context, usr user.User) (user.User, error) {
	if strings.TrimSpace(usr.Email) == "" {
		return user.User{}, user.ErrInvalidEmail
	}

	tx, err := r.dbc.BeginTx(ctx, nil)
	if err != nil {
		return user.User{}, err
	}

	createQuery, params, err := dialect.Insert(TABLE_USERS).Rows(
		goqu.Record{
			"name":  usr.Name,
			"email": usr.Email,
		}).Returning("created_at", "deleted_at", "email", "id", "name", "updated_at").ToSQL()
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

	query, params, err := dialect.From(TABLE_USERS).LeftOuterJoin(
		goqu.T(TABLE_METADATA),
		goqu.On(goqu.Ex{"users.id": goqu.I("metadata.user_id")})).Select("users.id", "name", "email", "key", "value", "users.created_at", "users.updated_at").Where(goqu.Or(
		goqu.C("name").ILike(fmt.Sprintf("%%%s%%", flt.Keyword)),
		goqu.C("email").ILike(fmt.Sprintf("%%%s%%", flt.Keyword)),
	)).Limit(uint(flt.Limit)).Offset(uint(offset)).ToSQL()
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

	query, params, err := dialect.From(TABLE_USERS).Select("id", "name", "email").Where(
		goqu.Ex{
			"id": goqu.Op{"in": userIDs},
		}).ToSQL()
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
		var transformedUser user.User
		transformedUser.ID = u.ID
		transformedUser.Email = u.Email
		transformedUser.Name = u.Name

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
				"updated_at": goqu.L("now()"),
			}).Where(
			goqu.Ex{
				"email": usr.Email,
			},
		).Returning("created_at", "deleted_at", "email", "id", "name", "updated_at").ToSQL()
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
	if strings.TrimSpace(usr.Email) == "" {
		return user.User{}, user.ErrInvalidEmail
	}

	var transformedUser user.User

	err := r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		query, params, err := dialect.Update(TABLE_USERS).Set(
			goqu.Record{
				"name":       usr.Name,
				"email":      usr.Email,
				"updated_at": goqu.L("now()"),
			}).Where(
			goqu.Ex{
				"id": usr.ID,
			},
		).Returning("created_at", "deleted_at", "email", "id", "name", "updated_at").ToSQL()
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

func (r UserRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	if strings.TrimSpace(email) == "" {
		return user.User{}, user.ErrInvalidEmail
	}

	var fetchedUser User
	data := make(map[string]any)

	query, params, err := dialect.From(TABLE_USERS).Where(
		goqu.Ex{
			"email": email,
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

func (r UserRepository) CreateMetadataKey(ctx context.Context, key user.UserMetadataKey) (user.UserMetadataKey, error) {
	if key.Key == "" {
		return user.UserMetadataKey{}, user.ErrEmptyKey
	}

	createQuery, params, err := dialect.Insert(TABLE_METADATA_KEYS).Rows(
		goqu.Record{
			"key":         key.Key,
			"description": key.Description,
		}).Returning("key", "description", "created_at", "updated_at").ToSQL()
	if err != nil {
		return user.UserMetadataKey{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var metadataKey UserMetadataKey
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_METADATA_KEYS,
				Operation:  "Create",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return r.dbc.QueryRowxContext(ctx, createQuery, params...).
			StructScan(&metadataKey)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, errDuplicateKey):
			return user.UserMetadataKey{}, user.ErrKeyAlreadyExists
		default:
			return user.UserMetadataKey{}, err
		}
	}

	return metadataKey.tranformUserMetadataKey(), nil
}
