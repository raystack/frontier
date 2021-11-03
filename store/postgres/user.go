package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/odpf/shield/internal/user"
)

type User struct {
	Id        string    `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Metadata  []byte    `db:"metadata"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

const (
	getUsersQuery            = `SELECT id, name,  email, metadata, created_at, updated_at from users where id=$1;`
	createUserQuery          = `INSERT INTO users(name, email, metadata) values($1, $2, $3) RETURNING id, name, email, metadata, created_at, updated_at;`
	listUsersQuery           = `SELECT id, name, email, metadata, created_at, updated_at from users;`
	selectUserForUpdateQuery = `SELECT id, name, email, metadata, version, updated_at from users where id=$1;`
	updateUserQuery          = `UPDATE users set name = $2, email = $3, metadata = $4, updated_at = now() where id = $1 RETURNING id, name, email, metadata, created_at, updated_at;`
)

func (s Store) GetUser(ctx context.Context, id string) (user.User, error) {
	fetchedUser, err := s.selectUser(ctx, id, false, nil)
	return fetchedUser, err
}

func (s Store) selectUser(ctx context.Context, id string, forUpdate bool, txn *sqlx.Tx) (user.User, error) {
	var fetchedUser User

	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		if forUpdate {
			return txn.GetContext(ctx, &fetchedUser, selectUserForUpdateQuery, id)
		} else {
			return s.DB.GetContext(ctx, &fetchedUser, getUsersQuery, id)
		}
	})

	if errors.Is(err, sql.ErrNoRows) {
		return user.User{}, user.UserDoesntExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return user.User{}, user.InvalidUUID
	} else if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedUser, err := transformToUser(fetchedUser)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedUser, nil
}

func (s Store) CreateUser(ctx context.Context, userToCreate user.User) (user.User, error) {
	marshaledMetadata, err := json.Marshal(userToCreate.Metadata)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var newUser User
	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newUser, createUserQuery, userToCreate.Name, userToCreate.Email, marshaledMetadata)
	})

	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedUser, err := transformToUser(newUser)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedUser, nil
}

func (s Store) ListUsers(ctx context.Context) ([]user.User, error) {
	var fetchedUsers []User
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedUsers, listUsersQuery)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []user.User{}, user.UserDoesntExist
	}

	if err != nil {
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

func (s Store) UpdateUser(ctx context.Context, toUpdate user.User) (user.User, error) {
	var updatedUser User

	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &updatedUser, updateUserQuery, toUpdate.Id, toUpdate.Name, toUpdate.Email, marshaledMetadata)
	})

	if err != nil {
		return user.User{}, fmt.Errorf("%s: %w", txnErr, err)
	}

	transformedUser, err := transformToUser(updatedUser)
	if err != nil {
		return user.User{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return transformedUser, nil
}

func transformToUser(from User) (user.User, error) {
	var unmarshalledMetadata map[string]string
	if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
		return user.User{}, err
	}

	return user.User{
		Id:        from.Id,
		Name:      from.Name,
		Email:     from.Email,
		Metadata:  unmarshalledMetadata,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}, nil
}
