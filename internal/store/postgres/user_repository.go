package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/db"
)

type UserRepository struct {
	dbc *db.Client
}

func NewUserRepository(dbc *db.Client) *UserRepository {
	return &UserRepository{
		dbc: dbc,
	}
}

func (r UserRepository) GetByID(ctx context.Context, id string) (user.User, error) {
	if id == "" {
		return user.User{}, user.ErrInvalidID
	}
	fetchedUser, err := r.selectUser(ctx, id, false, nil)
	return fetchedUser, err
}

func (r UserRepository) selectUser(ctx context.Context, id string, forUpdate bool, txn *sqlx.Tx) (user.User, error) {
	var fetchedUser User

	//TODO to be confirmed
	// selectUserForUpdateQuery, err := buildSelectUserForUpdateQuery(dialect)
	// if err != nil {
	// 	return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	// }

	query, params, err := dialect.From(TABLE_USERS).Select(&User{}).
		Where(goqu.Ex{
			"id": id,
		}).ToSQL()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		// if forUpdate {
		// 	return txn.GetContext(ctx, &fetchedUser, selectUserForUpdateQuery, id)
		// } else {
		return r.dbc.GetContext(ctx, &fetchedUser, query, params...)
		// }
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return user.User{}, user.ErrNotExist
		}
		if errors.Is(err, errInvalidTexRepresentation) {
			return user.User{}, user.ErrNotUUID
		}
		return user.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedUser, err := transformToUser(fetchedUser)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedUser, nil
}

func (r UserRepository) Create(ctx context.Context, usr user.User) (user.User, error) {
	marshaledMetadata, err := json.Marshal(usr.Metadata)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	query, params, err := dialect.Insert(TABLE_USERS).Rows(
		goqu.Record{
			"name":     usr.Name,
			"email":    usr.Email,
			"metadata": marshaledMetadata,
		}).Returning(&User{}).ToSQL()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var userModel User
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&userModel)
	}); err != nil {
		return user.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedUser, err := transformToUser(userModel)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedUser, nil
}

func (r UserRepository) List(ctx context.Context, flt user.Filter) ([]user.User, error) {
	var fetchedUsers []User

	var defaultLimit int32 = 50
	var defaultPage int32 = 1
	if flt.Limit < 1 {
		flt.Limit = defaultLimit
	}
	if flt.Page < 1 {
		flt.Page = defaultPage
	}

	offset := (flt.Page - 1) * flt.Limit

	query, params, err := dialect.From(TABLE_USERS).Where(goqu.Or(
		goqu.C("name").ILike(fmt.Sprintf("%%%s%%", flt.Keyword)),
		goqu.C("email").ILike(fmt.Sprintf("%%%s%%", flt.Keyword)),
	)).Limit(uint(flt.Limit)).Offset(uint(offset)).ToSQL()
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedUsers, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []user.User{}, nil
		}
		return []user.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedUsers []user.User
	for _, u := range fetchedUsers {
		transformedUser, err := transformToUser(u)
		if err != nil {
			return []user.User{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedUsers = append(transformedUsers, transformedUser)
	}

	return transformedUsers, nil
}

func (r UserRepository) GetByIDs(ctx context.Context, userIDs []string) ([]user.User, error) {
	var fetchedUsers []User

	query, params, err := dialect.From(TABLE_USERS).Select(&User{}).Where(
		goqu.Ex{
			"id": goqu.Op{"in": userIDs},
		}).ToSQL()
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedUsers, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return []user.User{}, user.ErrNotExist
		}
		if errors.Is(err, errInvalidTexRepresentation) {
			return []user.User{}, user.ErrNotUUID
		}
		return []user.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedUsers []user.User
	for _, u := range fetchedUsers {
		transformedUser, err := transformToUser(u)
		if err != nil {
			return []user.User{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedUsers = append(transformedUsers, transformedUser)
	}

	return transformedUsers, nil
}

func (r UserRepository) UpdateByEmail(ctx context.Context, usr user.User) (user.User, error) {
	if usr.Email == "" {
		return user.User{}, user.ErrInvalidEmail
	}

	marshaledMetadata, err := json.Marshal(usr.Metadata)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	query, params, err := dialect.Update(TABLE_USERS).Set(
		goqu.Record{
			"name":       usr.Name,
			"metadata":   marshaledMetadata,
			"updated_at": goqu.L("now()"),
		}).Where(
		goqu.Ex{
			"email": usr.Email,
		},
	).Returning(&User{}).ToSQL()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var userModel User
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&userModel)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user.User{}, user.ErrNotExist
		}
		return user.User{}, fmt.Errorf("%s: %w", txnErr, err)
	}

	transformedUser, err := transformToUser(userModel)
	if err != nil {
		return user.User{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return transformedUser, nil
}

func (r UserRepository) UpdateByID(ctx context.Context, usr user.User) (user.User, error) {
	if usr.ID == "" {
		return user.User{}, user.ErrInvalidID
	}

	marshaledMetadata, err := json.Marshal(usr.Metadata)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	query, params, err := dialect.Update(TABLE_USERS).Set(
		goqu.Record{
			"name":       usr.Name,
			"email":      usr.Email,
			"metadata":   marshaledMetadata,
			"updated_at": goqu.L("now()"),
		}).Where(
		goqu.Ex{
			"id": usr.ID,
		},
	).Returning(&User{}).ToSQL()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var userModel User
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&userModel)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user.User{}, user.ErrNotExist
		}
		return user.User{}, fmt.Errorf("%s: %w", txnErr, err)
	}

	transformedUser, err := transformToUser(userModel)
	if err != nil {
		return user.User{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return transformedUser, nil
}

func (r UserRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	if email == "" {
		return user.User{}, user.ErrInvalidEmail
	}

	query, params, err := dialect.From(TABLE_USERS).Where(
		goqu.Ex{
			"email": email,
		}).ToSQL()

	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var userModel User
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &userModel, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user.User{}, user.ErrNotExist
		}
		return user.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedUser, err := transformToUser(userModel)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedUser, nil
}
