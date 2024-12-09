package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/pkg/errors"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/db"
)

type UserRepository struct {
	dbc *db.Client
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
		return user.User{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_USERS, "GetByID", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &fetchedUser, userQuery, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, ErrDuplicateKey):
			return user.User{}, user.ErrConflict
		case errors.Is(err, sql.ErrNoRows):
			return user.User{}, user.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
			return user.User{}, user.ErrInvalidUUID
		default:
			return user.User{}, err
		}
	}

	transformedUser, err := fetchedUser.transformToUser()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	return transformedUser, nil
}

func (r UserRepository) GetByName(ctx context.Context, name string) (user.User, error) {
	if strings.TrimSpace(name) == "" {
		return user.User{}, user.ErrMissingName
	}

	var fetchedUser User
	query, params, err := dialect.From(TABLE_USERS).
		Where(goqu.Ex{
			"name": strings.ToLower(name),
		}).ToSQL()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_USERS, "GetByName", func(ctx context.Context) error {
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

	transformedUser, err := fetchedUser.transformToUser()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	return transformedUser, nil
}

func (r UserRepository) Create(ctx context.Context, usr user.User) (user.User, error) {
	if strings.TrimSpace(usr.Email) == "" || strings.TrimSpace(usr.Name) == "" {
		return user.User{}, user.ErrInvalidDetails
	}

	insertRow := goqu.Record{
		"name":       strings.ToLower(usr.Name),
		"email":      strings.ToLower(usr.Email),
		"title":      usr.Title,
		"avatar":     usr.Avatar,
		"created_at": goqu.L("now()"),
		"updated_at": goqu.L("now()"),
	}
	if usr.Metadata != nil {
		marshaledMetadata, err := json.Marshal(usr.Metadata)
		if err != nil {
			return user.User{}, fmt.Errorf("%w: %w", parseErr, err)
		}
		insertRow["metadata"] = marshaledMetadata
	}
	if usr.State != "" {
		insertRow["state"] = usr.State
	}
	createQuery, params, err := dialect.Insert(TABLE_USERS).Rows(insertRow).Returning(&User{}).ToSQL()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	tx, err := r.dbc.BeginTxx(ctx, nil)
	if err != nil {
		return user.User{}, err
	}

	var userModel User
	if err = r.dbc.WithTimeout(ctx, TABLE_USERS, "Create", func(ctx context.Context) error {
		return tx.QueryRowxContext(ctx, createQuery, params...).
			StructScan(&userModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, ErrDuplicateKey):
			return user.User{}, user.ErrConflict
		default:
			if err := tx.Rollback(); err != nil {
				return user.User{}, err
			}
			return user.User{}, err
		}
	}

	if err = tx.Commit(); err != nil {
		return user.User{}, err
	}

	transformedUser, err := userModel.transformToUser()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %w", parseErr, err)
	}
	return transformedUser, nil
}

func (r UserRepository) List(ctx context.Context, flt user.Filter) ([]user.User, error) {
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
		Select("users.id", "name", "email", "title", "avatar", "users.created_at", "users.updated_at")

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
		return []user.User{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var dbUsers []User
	if err = r.dbc.WithTimeout(ctx, TABLE_USERS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &dbUsers, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []user.User{}, nil
		}
		return []user.User{}, fmt.Errorf("%w: %w", dbErr, err)
	}

	var transformedUsers []user.User
	for _, dbUser := range dbUsers {
		u, err := dbUser.transformToUser()
		if err != nil {
			return nil, err
		}
		transformedUsers = append(transformedUsers, u)
	}

	return transformedUsers, nil
}

func (r UserRepository) GetByIDs(ctx context.Context, userIDs []string) ([]user.User, error) {
	if len(userIDs) == 0 {
		return []user.User{}, nil
	}
	var fetchedUsers []User

	query, params, err := dialect.From(TABLE_USERS).Select("id", "name", "email", "title", "avatar", "state").Where(
		goqu.Ex{
			"id": goqu.Op{"in": userIDs},
		}).Where(notDisabledUserExp).ToSQL()
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_USERS, "GetByIDs", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedUsers, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return []user.User{}, user.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
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
	if strings.TrimSpace(usr.Email) == "" {
		return user.User{}, user.ErrInvalidEmail
	}
	marshaledMetadata, err := json.Marshal(usr.Metadata)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %w", parseErr, err)
	}
	var transformedUser user.User

	err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		updateQuery, params, err := dialect.Update(TABLE_USERS).Set(
			goqu.Record{
				"title":      usr.Title,
				"avatar":     usr.Avatar,
				"metadata":   marshaledMetadata,
				"updated_at": goqu.L("now()"),
			}).Where(
			goqu.Ex{
				"email": strings.ToLower(usr.Email),
			},
		).Returning(&User{}).ToSQL()
		if err != nil {
			return fmt.Errorf("%w: %s", queryErr, err)
		}

		var userModel User
		if err = r.dbc.WithTimeout(ctx, TABLE_USERS, "UpdateByEmail", func(ctx context.Context) error {
			return tx.QueryRowxContext(ctx, updateQuery, params...).StructScan(&userModel)
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return user.ErrNotExist
			}
			return fmt.Errorf("%w: %w", txnErr, err)
		}

		transformedUser, err = userModel.transformToUser()
		if err != nil {
			return fmt.Errorf("%w: %w", parseErr, err)
		}

		return nil
	})

	if err != nil {
		return user.User{}, err
	}

	return transformedUser, nil
}

func (r UserRepository) UpdateByID(ctx context.Context, usr user.User) (user.User, error) {
	if usr.ID == "" || !utils.IsValidUUID(usr.ID) {
		return user.User{}, user.ErrInvalidID
	}

	marshaledMetadata, err := json.Marshal(usr.Metadata)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}
	var transformedUser user.User
	err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		query, params, err := dialect.Update(TABLE_USERS).Set(
			goqu.Record{
				"title":      usr.Title,
				"avatar":     usr.Avatar,
				"metadata":   marshaledMetadata,
				"updated_at": goqu.L("now()"),
			}).Where(
			goqu.Ex{
				"id": usr.ID,
			},
		).Returning(&User{}).ToSQL()
		if err != nil {
			return fmt.Errorf("%w: %s", queryErr, err)
		}

		var userModel User
		if err = r.dbc.WithTimeout(ctx, TABLE_USERS, "Update", func(ctx context.Context) error {
			return tx.QueryRowxContext(ctx, query, params...).StructScan(&userModel)
		}); err != nil {
			err = checkPostgresError(err)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return user.ErrNotExist
			case errors.Is(err, ErrDuplicateKey):
				return user.ErrConflict
			default:
				return err
			}
		}

		transformedUser, err = userModel.transformToUser()
		if err != nil {
			return fmt.Errorf("%w: %w", parseErr, err)
		}

		return nil
	})

	if err != nil {
		return user.User{}, err
	}

	return transformedUser, nil
}

func (r UserRepository) UpdateByName(ctx context.Context, usr user.User) (user.User, error) {
	if usr.Name == "" {
		return user.User{}, user.ErrMissingName
	}

	updateRecord := goqu.Record{
		"title":      usr.Title,
		"avatar":     usr.Avatar,
		"updated_at": goqu.L("now()"),
	}
	if len(usr.Metadata) > 0 {
		marshaledMetadata, err := json.Marshal(usr.Metadata)
		if err != nil {
			return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
		}
		updateRecord["metadata"] = marshaledMetadata
	}

	query, params, err := dialect.Update(TABLE_USERS).Set(updateRecord).Where(
		goqu.Ex{
			"name": strings.ToLower(usr.Name),
		},
	).Returning(&User{}).ToSQL()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var userModel User
	err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		if err = r.dbc.WithTimeout(ctx, TABLE_USERS, "UpdateByName", func(ctx context.Context) error {
			return tx.QueryRowxContext(ctx, query, params...).StructScan(&userModel)
		}); err != nil {
			err = checkPostgresError(err)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return user.ErrNotExist
			case errors.Is(err, ErrDuplicateKey):
				return user.ErrConflict
			default:
				return err
			}
		}

		return nil
	})
	if err != nil {
		return user.User{}, err
	}

	return userModel.transformToUser()
}

func (r UserRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	if strings.TrimSpace(email) == "" {
		return user.User{}, user.ErrInvalidEmail
	}

	var fetchedUser User
	query, params, err := dialect.From(TABLE_USERS).Where(
		goqu.Ex{
			"email": strings.ToLower(email),
		}).Where(notDisabledUserExp).ToSQL()

	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_USERS, "GetByEmail", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &fetchedUser, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user.User{}, user.ErrNotExist
		}
		return user.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedUser, err := fetchedUser.transformToUser()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

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
	).Returning(&User{}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	var userModel User
	if err = r.dbc.WithTimeout(ctx, TABLE_USERS, "SetState", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&userModel)
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
	).Returning(&User{}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	var userModel User
	if err = r.dbc.WithTimeout(ctx, TABLE_USERS, "Delete", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&userModel)
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
