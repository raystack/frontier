package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"

	"github.com/odpf/shield/internal/bootstrap/definition"
	"github.com/odpf/shield/internal/group"
	"github.com/odpf/shield/internal/user"
	"github.com/odpf/shield/model"
)

type User struct {
	Id        string       `db:"id"`
	Name      string       `db:"name"`
	Email     string       `db:"email"`
	Metadata  []byte       `db:"metadata"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

func listUserQueryHelper(page int32, limit int32) (uint, uint) {
	var defaultLimit int32 = 50
	if limit < 1 {
		limit = defaultLimit
	}

	offset := (page - 1) * limit

	return uint(limit), uint(offset)
}

func buildGetUserQuery(dialect goqu.DialectWrapper) (string, error) {
	getUserQuery, _, err := dialect.From("users").
		Where(goqu.Ex{
			"id": goqu.L("$1"),
		}).ToSQL()

	return getUserQuery, err
}

func buildGetUsersByIdsQuery(dialect goqu.DialectWrapper) (string, error) {
	getUsersByIdsQuery, _, err := dialect.From("users").Prepared(true).Where(
		goqu.C("id").In("id_PH")).ToSQL()

	return getUsersByIdsQuery, err
}

func buildGetCurrentUserQuery(dialect goqu.DialectWrapper) (string, error) {
	getCurrentUserQuery, _, err := dialect.From("users").Where(
		goqu.Ex{
			"email": goqu.L("$1"),
		}).ToSQL()

	return getCurrentUserQuery, err
}

func buildCreateUserQuery(dialect goqu.DialectWrapper) (string, error) {
	createUserQuery, _, err := dialect.Insert("users").Rows(
		goqu.Record{
			"name":     goqu.L("$1"),
			"email":    goqu.L("$2"),
			"metadata": goqu.L("$3"),
		}).Returning(&User{}).ToSQL()

	return createUserQuery, err
}

func buildListUsersQuery(dialect goqu.DialectWrapper, limit int32, page int32, keyword string) (string, error) {
	limitRows, offset := listUserQueryHelper(page, limit)

	listUsersQuery, _, err := dialect.From("users").Where(goqu.Or(
		goqu.C("name").ILike(fmt.Sprintf("%%%s%%", keyword)),
		goqu.C("email").ILike(fmt.Sprintf("%%%s%%", keyword)),
	)).Limit(limitRows).Offset(offset).ToSQL()

	return listUsersQuery, err
}

func buildSelectUserForUpdateQuery(dialect goqu.DialectWrapper) (string, error) {
	selectUserForUpdateQuery, _, err := dialect.From("users").Prepared(true).Where(
		goqu.Ex{
			"id": "id_PH",
		}).ToSQL()

	return selectUserForUpdateQuery, err
}

func buildUpdateUserQuery(dialect goqu.DialectWrapper) (string, error) {
	updateUserQuery, _, err := dialect.Update("users").Set(
		goqu.Record{
			"name":       goqu.L("$2"),
			"email":      goqu.L("$3"),
			"metadata":   goqu.L("$4"),
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{
		"id": goqu.L("$1"),
	}).Returning(&User{}).ToSQL()

	return updateUserQuery, err
}

func buildUpdateCurrentUserQuery(dialect goqu.DialectWrapper) (string, error) {
	updateCurrentUserQuery, _, err := dialect.Update("users").Set(
		goqu.Record{
			"name":       goqu.L("$2"),
			"metadata":   goqu.L("$3"),
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{
		"email": goqu.L("$1"),
	}).Returning(&User{}).ToSQL()

	return updateCurrentUserQuery, err
}

func buildListUserGroupsQuery(dialect goqu.DialectWrapper) (string, error) {
	listUserGroupsQuery, _, err := dialect.Select(
		goqu.I("g.id").As("id"),
		goqu.I("g.metadata").As("metadata"),
		goqu.I("g.name").As("name"),
		goqu.I("g.slug").As("slug"),
		goqu.I("g.updated_at").As("updated_at"),
		goqu.I("g.created_at").As("created_at"),
		goqu.I("g.org_id").As("org_id"),
	).From(goqu.L("relations r")).
		Join(goqu.L("groups g"), goqu.On(
			goqu.I("g.id").Cast("VARCHAR").
				Eq(goqu.I("r.object_id")),
		)).Where(goqu.Ex{
		"r.object_namespace_id": definition.TeamNamespace.Id,
		"subject_namespace_id":  definition.UserNamespace.Id,
		"subject_id":            goqu.L("$1"),
		"role_id":               goqu.L("$2"),
	}).ToSQL()

	return listUserGroupsQuery, err
}

func (s Store) GetUser(ctx context.Context, id string) (model.User, error) {
	fetchedUser, err := s.selectUser(ctx, id, false, nil)
	return fetchedUser, err
}

func (s Store) selectUser(ctx context.Context, id string, forUpdate bool, txn *sqlx.Tx) (model.User, error) {
	var fetchedUser User

	selectUserForUpdateQuery, err := buildSelectUserForUpdateQuery(dialect)
	if err != nil {
		return model.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var getUserQuery string
	getUserQuery, err = buildGetUserQuery(dialect)
	if err != nil {
		return model.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
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
	createUserQuery, err := buildCreateUserQuery(dialect)
	if err != nil {
		return model.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

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

func (s Store) ListUsers(ctx context.Context, limit int32, page int32, keyword string) (model.PagedUsers, error) {
	var fetchedUsers []User
	listUsersQuery, err := buildListUsersQuery(dialect, limit, page, keyword)
	if err != nil {
		return model.PagedUsers{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedUsers, listUsersQuery)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return model.PagedUsers{}, user.UserDoesntExist
	}

	if err != nil {
		return model.PagedUsers{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedUsers []model.User

	for _, u := range fetchedUsers {
		transformedUser, err := transformToUser(u)
		if err != nil {
			return model.PagedUsers{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedUsers = append(transformedUsers, transformedUser)
	}

	res := model.PagedUsers{
		Count: int32(len(fetchedUsers)),
		Users: transformedUsers,
	}

	return res, nil
}

func (s Store) GetUsersByIds(ctx context.Context, userIds []string) ([]model.User, error) {
	var fetchedUsers []User

	getUsersByIdsQuery, err := buildGetUsersByIdsQuery(dialect)
	if err != nil {
		return []model.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var query string
	var args []interface{}
	query, args, err = sqlx.In(getUsersByIdsQuery, userIds)

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

	var updateUserQuery string
	updateUserQuery, err = buildUpdateUserQuery(dialect)
	if err != nil {
		return model.User{}, fmt.Errorf("%w: %s", queryErr, err)
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

	getCurrentUserQuery, err := buildGetCurrentUserQuery(dialect)
	if err != nil {
		return model.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
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

	var updateCurrentUserQuery string
	updateCurrentUserQuery, err = buildUpdateCurrentUserQuery(dialect)
	if err != nil {
		return model.User{}, fmt.Errorf("%w: %s", queryErr, err)
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

func (s Store) ListUserGroups(ctx context.Context, userId string, roleId string) ([]model.Group, error) {
	role := definition.TeamMemberRole.Id

	if roleId == definition.TeamAdminRole.Id {
		role = definition.TeamAdminRole.Id
	}

	var fetchedGroups []Group

	listUserGroupsQuery, err := buildListUserGroupsQuery(dialect)
	if err != nil {
		return []model.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedGroups, listUserGroupsQuery, userId, role)
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
