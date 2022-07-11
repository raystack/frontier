package postgres

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"strings"
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

const (
	UUID_MIN        = "00000000-0000-0000-0000-000000000000"
	UUID_MAX        = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	PAGE_NEXT       = "next"
	PAGE_PREV       = "prev"
	TOKEN_SEPARATOR = "__"
)

func PageTokenizer(label string, uuid string) string {
	if uuid == "" {
		uuid = UUID_MIN
	}
	data := fmt.Sprintf("%s%s%s", label, TOKEN_SEPARATOR, uuid)
	encoded := base64.StdEncoding.EncodeToString([]byte(data))

	return encoded
}

func PageDetokenizer(token string) (string, string, error) {
	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", "", err
	}

	dataStr := string(data)
	dataList := strings.Split(dataStr, TOKEN_SEPARATOR)
	label, uuid := dataList[0], dataList[1]

	return label, uuid, err
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

func buildListUsersQuery(dialect goqu.DialectWrapper, limit int32, pageToken string, keyword string) (string, error) {
	label, uuid, err := PageDetokenizer(pageToken)
	if err != nil {
		return "", err
	}

	var listUsersQuery string
	if label == PAGE_NEXT {
		// Next Page Query
		listUsersQuery, _, err = dialect.From("users").Where(goqu.And(
			goqu.Or(
				goqu.C("name").ILike(fmt.Sprintf("%%%s%%", keyword)),
				goqu.C("email").ILike(fmt.Sprintf("%%%s%%", keyword)),
			), goqu.C("id").Gt(uuid),
		)).Order(goqu.C("id").Asc()).Limit(uint(limit)).ToSQL()
	} else {
		// Previous Page Query
		listUsersInnerQuery := dialect.From("users").Where(goqu.And(
			goqu.Or(
				goqu.C("name").ILike(fmt.Sprintf("%%%s%%", keyword)),
				goqu.C("email").ILike(fmt.Sprintf("%%%s%%", keyword)),
			), goqu.C("id").Lt(uuid),
		)).Order(goqu.C("id").Desc()).Limit(uint(limit))

		listUsersQuery, _, err = goqu.From(listUsersInnerQuery).Order(goqu.C("id").Asc()).ToSQL()
	}

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

// Helper Function
func (s Store) getLimits(ctx context.Context, label string, uuid string, keyword string) (string, error) {
	var fetchedUser User

	pageToken := PageTokenizer(label, uuid)
	query, err := buildListUsersQuery(dialect, 1, pageToken, keyword)
	if err != nil {
		return "", fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedUser, query)
	})

	var labelId string
	if errors.Is(err, sql.ErrNoRows) {
		if label == PAGE_NEXT {
			labelId = UUID_MAX
		} else if label == PAGE_PREV {
			labelId = UUID_MIN
		}
	} else if err != nil {
		return "", fmt.Errorf("%w: %s", dbErr, err)
	}

	if err == nil {
		labelId = uuid
	}

	return labelId, nil
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

func (s Store) ListUsers(ctx context.Context, limit int32, pageToken string, keyword string) (model.PagedUser, error) {
	var fetchedUsers []User

	if pageToken == "" {
		pageToken = PageTokenizer(PAGE_NEXT, UUID_MIN)
	} else if pageToken == PageTokenizer(PAGE_PREV, UUID_MIN) {
		return model.PagedUser{
			Count:             0,
			PreviousPageToken: pageToken,
			NextPageToken:     PageTokenizer(PAGE_NEXT, UUID_MIN),
			Users:             []model.User{},
		}, nil
	} else if pageToken == PageTokenizer(PAGE_NEXT, UUID_MAX) {
		return model.PagedUser{
			Count:             0,
			PreviousPageToken: PageTokenizer(PAGE_PREV, UUID_MAX),
			NextPageToken:     pageToken,
			Users:             []model.User{},
		}, nil
	}

	listUsersQuery, err := buildListUsersQuery(dialect, limit, pageToken, keyword)
	if err != nil {
		return model.PagedUser{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedUsers, listUsersQuery)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return model.PagedUser{}, user.UserDoesntExist
	}

	if err != nil {
		return model.PagedUser{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedUsers []model.User
	var nextToken, prevToken string

	lenUsers := len(fetchedUsers)
	for idx, u := range fetchedUsers {
		transformedUser, err := transformToUser(u)
		if err != nil {
			return model.PagedUser{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedUsers = append(transformedUsers, transformedUser)
		if idx == 0 {
			prevId, err := s.getLimits(ctx, PAGE_PREV, transformedUser.Id, keyword)
			if err != nil {
				return model.PagedUser{}, err
			}
			prevToken = PageTokenizer(PAGE_PREV, prevId)
		}
		if idx == (lenUsers - 1) {
			nextId, err := s.getLimits(ctx, PAGE_NEXT, transformedUser.Id, keyword)
			if err != nil {
				return model.PagedUser{}, err
			}
			nextToken = PageTokenizer(PAGE_NEXT, nextId)
		}
	}

	res := model.PagedUser{
		Users:             transformedUsers,
		PreviousPageToken: prevToken,
		NextPageToken:     nextToken,
		Count:             int32(lenUsers),
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
