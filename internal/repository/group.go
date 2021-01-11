package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/x1unix/sbda-ledger/internal/model/user"
)

const (
	groupsTable       = "groups"
	groupMembersTable = "group_membership"

	colGroupID  = "group_id"
	colMemberID = "member_id"
	colOwnerID  = "owner_id"
)

var (
	groupCols = []string{colID, colName, colOwnerID}
)

type GroupRepository struct {
	db *sqlx.DB
}

// NewGroupRepository is GroupRepository constructor
func NewGroupRepository(db *sqlx.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r GroupRepository) AddGroup(ctx context.Context, g user.GroupInfo) (*user.GroupID, error) {
	q, args, err := psql.Insert(groupsTable).SetMap(map[string]interface{}{
		colName:    g.Name,
		colOwnerID: g.OwnerID,
	}).Suffix(returnIDSuffix).ToSql()
	if err != nil {
		return nil, err
	}

	retId := new(user.GroupID)
	err = r.db.GetContext(ctx, retId, q, args...)
	return retId, err
}

func (r GroupRepository) DeleteGroup(ctx context.Context, gid user.GroupID) error {
	result, err := psql.Delete(groupsTable).Where(squirrel.Eq{colID: gid}).
		RunWith(r.db).ExecContext(ctx)

	if err != nil {
		return err
	}

	return checkAffectedRows(result)
}
