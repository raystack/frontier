package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/odpf/shield/internal/bootstrap/definition"
	"github.com/odpf/shield/internal/group"

	"github.com/odpf/shield/internal/user"
	"github.com/odpf/shield/model"

	"github.com/jmoiron/sqlx"
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
	getUserQuery             = `SELECT id, name,  email, metadata, created_at, updated_at from users where id=$1;`
	getUsersByIdsQuery       = `SELECT id, name,  email, metadata, created_at, updated_at from users where id IN (?);`
	getCurrentUserQuery      = `SELECT id, name, email, metadata, created_at, updated_at from users where email=$1;`
	createUserQuery          = `INSERT INTO users(name, email, metadata) values($1, $2, $3) RETURNING id, name, email, metadata, created_at, updated_at;`
	listUsersQuery           = `SELECT id, name, email, metadata, created_at, updated_at from users;`
	selectUserForUpdateQuery = `SELECT id, name, email, metadata, updated_at from users where id=$1;`
	updateUserQuery          = `UPDATE users set name = $2, email = $3, metadata = $4, updated_at = now() where id = $1 RETURNING id, name, email, metadata, created_at, updated_at;`
	updateCurrentUserQuery   = `UPDATE users set name = $2, metadata = $3, updated_at = now() where email = $1 RETURNING id, name, email, metadata, created_at, updated_at;`
)

var (
	listUserGroupsQuery = fmt.Sprintf(
		`SELECT g.id as id, g.metadata as metadata, g."name" as "name", g.slug as slug, g.updated_at as updated_at, g.created_at as created_at, g.org_id as org_id 
				FROM relations r 
				JOIN groups g ON CAST(g.id as VARCHAR) = r.object_id
				WHERE r.object_namespace_id = '%s'
					AND subject_namespace_id = '%s'
					AND subject_id = $1
					AND role_id = $2;`,
		definition.TeamNamespace.Id, definition.UserNamespace.Id,
	)
)

func (s Store) GetUser(ctx context.Context, id string) (model.User, error) {
	fetchedUser, err := s.selectUser(ctx, id, false, nil)
	return fetchedUser, err
}

func (s Store) selectUser(ctx context.Context, id string, forUpdate bool, txn *sqlx.Tx) (model.User, error) {
	var fetchedUser User

	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		if forUpdate {
			return txn.GetContext(ctx, &fetchedUser, selectUserForUpdateQuery, id)
		} else {
			return s.DB.GetContext(ctx, &fetchedUser, getUserQuery, id)
		}
	})

	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, user.UserDoesntExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return model.User{}, user.InvalidUUID
	} else if err != nil {
		return model.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedUser, err := transformToUser(fetchedUser)
	if err != nil {
		return model.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedUser, nil
}

func (s Store) CreateUser(ctx context.Context, userToCreate model.User) (model.User, error) {
	marshaledMetadata, err := json.Marshal(userToCreate.Metadata)
	if err != nil {
		return model.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var newUser User
	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newUser, createUserQuery, userToCreate.Name, userToCreate.Email, marshaledMetadata)
	})

	if err != nil {
		return model.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedUser, err := transformToUser(newUser)
	if err != nil {
		return model.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedUser, nil
}

func (s Store) ListUsers(ctx context.Context) ([]model.User, error) {
	var fetchedUsers []User
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedUsers, listUsersQuery)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []model.User{}, user.UserDoesntExist
	}

	if err != nil {
		return []model.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedUsers []model.User

	for _, u := range fetchedUsers {
		transformedUser, err := transformToUser(u)
		if err != nil {
			return []model.User{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedUsers = append(transformedUsers, transformedUser)
	}

	return transformedUsers, nil
}

func (s Store) GetUsersByIds(ctx context.Context, userIds []string) ([]model.User, error) {
	var fetchedUsers []User

	query, args, err := sqlx.In(getUsersByIdsQuery, userIds)

	if err != nil {
		return []model.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	query = s.DB.Rebind(query)

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedUsers, query, args...)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []model.User{}, user.UserDoesntExist
	}

	if err != nil {
		return []model.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedUsers []model.User

	for _, u := range fetchedUsers {
		transformedUser, err := transformToUser(u)
		if err != nil {
			return []model.User{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedUsers = append(transformedUsers, transformedUser)
	}

	return transformedUsers, nil
}

func (s Store) UpdateUser(ctx context.Context, toUpdate model.User) (model.User, error) {
	var updatedUser User

	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return model.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &updatedUser, updateUserQuery, toUpdate.Id, toUpdate.Name, toUpdate.Email, marshaledMetadata)
	})

	if err != nil {
		return model.User{}, fmt.Errorf("%s: %w", txnErr, err)
	}

	transformedUser, err := transformToUser(updatedUser)
	if err != nil {
		return model.User{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return transformedUser, nil
}

func (s Store) GetCurrentUser(ctx context.Context, email string) (model.User, error) {
	self, err := s.getUserWithEmailID(ctx, email)
	return self, err
}

func (s Store) getUserWithEmailID(ctx context.Context, email string) (model.User, error) {
	var userSelf User
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &userSelf, getCurrentUserQuery, email)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, user.UserDoesntExist
	} else if err != nil {
		return model.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedUser, err := transformToUser(userSelf)
	if err != nil {
		return model.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedUser, nil
}

func (s Store) UpdateCurrentUser(ctx context.Context, toUpdate model.User) (model.User, error) {
	var updatedUser User

	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return model.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &updatedUser, updateCurrentUserQuery, toUpdate.Email, toUpdate.Name, marshaledMetadata)
	})

	if err != nil {
		return model.User{}, fmt.Errorf("%s: %w", txnErr, err)
	}

	transformedUser, err := transformToUser(updatedUser)
	if err != nil {
		return model.User{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return transformedUser, nil
}

func (s Store) ListUserGroups(ctx context.Context, userId string) ([]model.Group, error) {
	var fetchedGroups []Group
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedGroups, listUserGroupsQuery, userId, definition.TeamMemberRole.Id)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []model.Group{}, group.GroupDoesntExist
	}

	if err != nil {
		return []model.Group{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedGroups []model.Group

	for _, v := range fetchedGroups {
		transformedGroup, err := transformToGroup(v)
		if err != nil {
			return []model.Group{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedGroups = append(transformedGroups, transformedGroup)
	}

	return transformedGroups, nil
}

func transformToUser(from User) (model.User, error) {
	var unmarshalledMetadata map[string]string
	if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
		return model.User{}, err
	}

	return model.User{
		Id:        from.Id,
		Name:      from.Name,
		Email:     from.Email,
		Metadata:  unmarshalledMetadata,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}, nil
}
