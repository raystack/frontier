package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"database/sql"

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

	var getUserQuery string
	getUserQuery, err := buildGetUserQuery(dialect)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		// if forUpdate {
		// 	return txn.GetContext(ctx, &fetchedUser, selectUserForUpdateQuery, id)
		// } else {
		return r.dbc.GetContext(ctx, &fetchedUser, getUserQuery, id)
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

func (r UserRepository) Create(ctx context.Context, userToCreate user.User) (user.User, error) {
	marshaledMetadata, err := json.Marshal(userToCreate.Metadata)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	createUserQuery, err := buildCreateUserQuery(dialect)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var userModel User
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, createUserQuery, userToCreate.Name, userToCreate.Email, marshaledMetadata).StructScan(&userModel)
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
	listUsersQuery, err := buildListUsersQuery(dialect, flt.Limit, flt.Page, flt.Keyword)
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedUsers, listUsersQuery)
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

	getUsersByIDsQuery, err := buildGetUsersByIDsQuery(dialect)
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var query string
	var args []interface{}
	query, args, err = sqlx.In(getUsersByIDsQuery, userIDs)
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	query = r.dbc.Rebind(query)
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedUsers, query, args...)
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

func (r UserRepository) UpdateByEmail(ctx context.Context, toUpdate user.User) (user.User, error) {
	if toUpdate.Email == "" {
		return user.User{}, user.ErrInvalidEmail
	}

	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var updateUserByEmailQuery string
	updateUserByEmailQuery, err = buildUpdateUserByEmailQuery(dialect)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var userModel User
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, updateUserByEmailQuery, toUpdate.Email, toUpdate.Name, marshaledMetadata).StructScan(&userModel)
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

func (r UserRepository) UpdateByID(ctx context.Context, toUpdate user.User) (user.User, error) {
	if toUpdate.ID == "" {
		return user.User{}, user.ErrInvalidID
	}

	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var updateUserByIDQuery string
	updateUserByIDQuery, err = buildUpdateUserByIDQuery(dialect)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var userModel User
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, updateUserByIDQuery, toUpdate.ID, toUpdate.Name, toUpdate.Email, marshaledMetadata).StructScan(&userModel)
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

	getUserByEmailQuery, err := buildGetUserByEmailQuery(dialect)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var userModel User
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &userModel, getUserByEmailQuery, email)
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
