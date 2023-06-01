package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/odpf/shield/pkg/utils"

	"github.com/pkg/errors"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/db"
)

type UserRepository struct {
	dbc *db.Client
}

type joinUserMetadata struct {
	ID        string         `db:"id"`
	Name      string         `db:"name"`
	Email     string         `db:"email"`
	Title     sql.NullString `db:"title"`
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
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedUser, nil
}

func (r UserRepository) GetByName(ctx context.Context, name string) (user.User, error) {
	if strings.TrimSpace(name) == "" {
		return user.User{}, user.ErrMissingSlug
	}

	var fetchedUser User
	query, params, err := dialect.From(TABLE_USERS).
		Where(goqu.Ex{
			"name": name,
		}).ToSQL()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
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
		return user.User{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedUser, nil
}

func (r UserRepository) Create(ctx context.Context, usr user.User) (user.User, error) {
	if strings.TrimSpace(usr.Email) == "" || strings.TrimSpace(usr.Name) == "" {
		return user.User{}, user.ErrInvalidDetails
	}

	tx, err := r.dbc.BeginTx(ctx, nil)
	if err != nil {
		return user.User{}, err
	}

	insertRow := goqu.Record{
		"name":  usr.Name,
		"email": usr.Email,
		"title": usr.Title,
	}
	if usr.State != "" {
		insertRow["state"] = usr.State
	}
	createQuery, params, err := dialect.Insert(TABLE_USERS).Rows(insertRow).Returning("created_at", "deleted_at", "email", "id", "name", "title", "state", "updated_at").ToSQL()
	if err != nil {
		return user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var userModel User
	if err = r.dbc.WithTimeout(ctx, TABLE_USERS, "Upsert", func(ctx context.Context) error {
		return tx.QueryRowContext(ctx, createQuery, params...).
			Scan(&userModel.CreatedAt,
				&userModel.DeletedAt,
				&userModel.Email,
				&userModel.ID,
				&userModel.Name,
				&userModel.Title,
				&userModel.State,
				&userModel.UpdatedAt,
			)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, ErrDuplicateKey):
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

	err = tx.Commit()
	if err != nil {
		return user.User{}, err
	}

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
		Select("users.id", "name", "email", "title", "users.created_at", "users.updated_at")

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

	if err = r.dbc.WithTimeout(ctx, TABLE_USERS, "List", func(ctx context.Context) error {
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
		currentUser.Title = u.Title.String
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

	query, params, err := dialect.From(TABLE_USERS).Select("id", "name", "email", "title", "state").Where(
		goqu.Ex{
			"id": goqu.Op{"in": userIDs},
		}).Where(notDisabledUserExp).ToSQL()
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", queryErr, err)
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
	userMetadata := make(map[string]any)

	if strings.TrimSpace(usr.Email) == "" {
		return user.User{}, user.ErrInvalidEmail
	}

	var transformedUser user.User

	err := r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		updateQuery, params, err := dialect.Update(TABLE_USERS).Set(
			goqu.Record{
				"name":       usr.Name,
				"title":      usr.Title,
				"updated_at": goqu.L("now()"),
			}).Where(
			goqu.Ex{
				"email": usr.Email,
			},
		).Returning("created_at", "deleted_at", "email", "id", "name", "state", "title", "updated_at").ToSQL()
		if err != nil {
			return fmt.Errorf("%w: %s", queryErr, err)
		}

		var userModel User
		if err = r.dbc.WithTimeout(ctx, TABLE_USERS, "UpdateByEmail", func(ctx context.Context) error {
			return tx.QueryRowContext(ctx, updateQuery, params...).
				Scan(&userModel.CreatedAt,
					&userModel.DeletedAt,
					&userModel.Email,
					&userModel.ID,
					&userModel.Name,
					&userModel.State,
					&userModel.Title,
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

		return nil
	})

	if err != nil {
		return user.User{}, err
	}

	transformedUser.Metadata = userMetadata

	return transformedUser, nil
}

func (r UserRepository) UpdateByID(ctx context.Context, usr user.User) (user.User, error) {
	if usr.ID == "" || !utils.IsValidUUID(usr.ID) {
		return user.User{}, user.ErrInvalidID
	}
	if strings.TrimSpace(usr.Email) == "" || strings.TrimSpace(usr.Name) == "" {
		return user.User{}, user.ErrInvalidDetails
	}

	var transformedUser user.User

	err := r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		query, params, err := dialect.Update(TABLE_USERS).Set(
			goqu.Record{
				"name":       usr.Name,
				"title":      usr.Title,
				"updated_at": goqu.L("now()"),
			}).Where(
			goqu.Ex{
				"id": usr.ID,
			},
		).Returning("created_at", "deleted_at", "email", "id", "name", "state", "title", "updated_at").ToSQL()
		if err != nil {
			return fmt.Errorf("%w: %s", queryErr, err)
		}

		var userModel User
		if err = r.dbc.WithTimeout(ctx, TABLE_USERS, "Update", func(ctx context.Context) error {
			return tx.QueryRowContext(ctx, query, params...).Scan(&userModel.CreatedAt,
				&userModel.DeletedAt,
				&userModel.Email,
				&userModel.ID,
				&userModel.Name,
				&userModel.State,
				&userModel.Title,
				&userModel.UpdatedAt,
			)
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
			return fmt.Errorf("%s: %w", parseErr, err)
		}

		return nil
	})

	if err != nil {
		return user.User{}, err
	}

	transformedUser.Metadata = usr.Metadata
	return transformedUser, nil
}

func (r UserRepository) UpdateByName(ctx context.Context, usr user.User) (user.User, error) {
	if usr.Name == "" {
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
				"title":      usr.Name,
				"updated_at": goqu.L("now()"),
			}).Where(
			goqu.Ex{
				"slug": usr.ID,
			},
		).Returning("created_at", "deleted_at", "email", "id", "name", "title", "updated_at").ToSQL()
		if err != nil {
			return fmt.Errorf("%w: %s", queryErr, err)
		}

		var userModel User
		if err = r.dbc.WithTimeout(ctx, TABLE_USERS, "UpdateByName", func(ctx context.Context) error {
			return tx.QueryRowContext(ctx, query, params...).Scan(&userModel.CreatedAt,
				&userModel.DeletedAt,
				&userModel.Email,
				&userModel.ID,
				&userModel.Name,
				&userModel.Title,
				&userModel.UpdatedAt)
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
			return fmt.Errorf("%s: %w", parseErr, err)
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

	if err = r.dbc.WithTimeout(ctx, TABLE_USERS, "SetState", func(ctx context.Context) error {
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

	if err = r.dbc.WithTimeout(ctx, TABLE_USERS, "Delete", func(ctx context.Context) error {
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
