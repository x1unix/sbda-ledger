package repository

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/x1unix/sbda-ledger/internal/model/user"
	"github.com/x1unix/sbda-ledger/internal/service"
)

const (
	colID       = "id"
	colEmail    = "email"
	colName     = "name"
	colPassword = "password"

	tableUsers = "users"
)

var userCols = []string{colID, colEmail, colName, colPassword}

type UserRepository struct {
	db *sqlx.DB
}

func (r UserRepository) AddUser(ctx context.Context, u user.User) (*user.ID, error) {
	q, args, err := psql.Insert(tableUsers).SetMap(map[string]interface{}{
		colEmail:    u.Email,
		colName:     u.Name,
		colPassword: u.PasswordHash,
	}).Suffix("RETURNING " + colID).ToSql()
	if err != nil {
		return nil, err
	}

	newID := new(user.ID)
	return newID, r.db.GetContext(ctx, newID, q, args...)
}

func (r UserRepository) UserByEmail(email string) (*user.User, error) {
	q, args, err := psql.Select(userCols...).Where(squirrel.Eq{
		colEmail: email,
	}).Limit(1).ToSql()
	if err != nil {
		return nil, err
	}
	u := new(user.User)
	err = r.db.Get(u, q, args...)
	return u, wrapRecordError(err)
}

func (r UserRepository) Exists(email string) (bool, error) {
	q, args, err := psql.Select("COUNT(*)").Where(squirrel.Eq{
		colEmail: email,
	}).ToSql()
	if err != nil {
		return false, err
	}
	var count uint
	err = r.db.Get(&count, q, args...)
	return count > 0, err
}

func wrapRecordError(err error) error {
	switch err {
	case nil:
		return nil
	case sql.ErrNoRows:
		return service.ErrNotExists
	default:
		return err
	}
}
