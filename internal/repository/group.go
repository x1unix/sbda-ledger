package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/x1unix/sbda-ledger/internal/model/user"
	"github.com/x1unix/sbda-ledger/internal/service"
)

const (
	tableGroups       = "groups"
	tableGroupMembers = "group_membership"

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

// AddGroup implements service.GroupStore
func (r GroupRepository) AddGroup(ctx context.Context, name string, owner user.ID) (*user.GroupID, error) {
	q, args, err := psql.Insert(tableGroups).SetMap(map[string]interface{}{
		colName:    name,
		colOwnerID: owner,
	}).Suffix(returnIDSuffix).ToSql()
	if err != nil {
		return nil, err
	}

	retId := new(user.GroupID)
	err = r.db.GetContext(ctx, retId, q, args...)
	return retId, err
}

// GetGroupOwner implements service.GroupStore
func (r GroupRepository) GetGroupOwner(ctx context.Context, gid user.GroupID) (*user.ID, error) {
	q, args, err := psql.Select(colOwnerID).From(tableGroups).Where(squirrel.Eq{
		colID: gid,
	}).Limit(1).ToSql()
	if err != nil {
		return nil, err
	}

	uid := new(user.ID)
	err = r.db.GetContext(ctx, uid, q, args...)
	if err == sql.ErrNoRows {
		return nil, service.ErrGroupNotFound
	}

	return uid, err
}

// AddGroupUsers implements service.GroupManager
func (r GroupRepository) AddGroupUsers(ctx context.Context, gid user.GroupID, uids []user.ID) error {
	qb := psql.Insert(tableGroupMembers).Columns(colGroupID, colMemberID)
	for _, uid := range uids {
		qb = qb.Values(gid, uid)
	}

	_, err := qb.RunWith(r.db).ExecContext(ctx)
	return err
}

// DeleteGroupUser implements service.GroupManager
func (r GroupRepository) DeleteGroupUser(ctx context.Context, gid user.GroupID, uid user.ID) error {
	rows, err := psql.Delete(tableGroupMembers).Where(squirrel.Eq{
		colGroupID:  gid,
		colMemberID: uid,
	}).RunWith(r.db).ExecContext(ctx)
	if err != nil {
		return err
	}

	return checkAffectedRows(rows)
}

// DeleteGroup implements service.GroupStore
func (r GroupRepository) DeleteGroup(ctx context.Context, gid user.GroupID) error {
	result, err := psql.Delete(tableGroups).Where(squirrel.Eq{colID: gid}).
		RunWith(r.db).ExecContext(ctx)

	if err != nil {
		return err
	}

	return checkAffectedRows(result)
}

// GroupByID implements service.GroupStore
func (r GroupRepository) GroupByID(ctx context.Context, gid user.ID) (*user.Group, error) {
	q, args, err := psql.Select(groupCols...).From(tableGroups).
		Where(squirrel.Eq{colID: gid}).Limit(1).ToSql()
	if err != nil {
		return nil, err
	}

	out := new(user.Group)
	err = r.db.GetContext(ctx, out, q, args...)
	if err == sql.ErrNoRows {
		return nil, service.ErrGroupNotFound
	}

	return out, err
}

// GroupsByUser implements service.GroupManager
func (r GroupRepository) GroupsByUser(ctx context.Context, uid user.ID) (user.Groups, error) {
	// TODO: replace squirrel with gogu everywhere somewhere in future

	// squirrel doesn't support union selects still.
	// "github.com/doug-martin/goqu/v9" supports it, but
	// I don't want to bring a new lib just for 1 query.
	const query = "(SELECT id, name, owner_id FROM " + tableGroups + " WHERE owner_id = $1)" +
		" UNION " +
		"(SELECT " +
		"id, name, owner_id" +
		" FROM " + tableGroupMembers + " m" +
		" INNER JOIN " + tableGroups + " g on " +
		"m.group_id = g.id" +
		" WHERE " +
		"m.member_id = $1" +
		")"

	var out user.Groups
	err := r.db.SelectContext(ctx, &out, query, uid)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return out, err
}

func (r GroupRepository) GroupMemberIDs(ctx context.Context, gid user.GroupID) ([]user.ID, error) {
	const query = "(SELECT owner_id as uid from " + tableGroups + " WHERE id = $1)" +
		" UNION " +
		"(SELECT member_id as uid FROM " + tableGroupMembers + " WHERE group_id = $1)"
	var out []user.ID
	err := r.db.SelectContext(ctx, &out, query, gid)
	if err == sql.ErrNoRows {
		return nil, service.ErrGroupNotFound
	}

	return out, nil
}

// GroupsByAuthor implements service.GroupManager
func (r GroupRepository) GroupsByAuthor(ctx context.Context, uid user.ID) (user.Groups, error) {
	q, args, err := psql.Select(groupCols...).From(tableGroups).
		Where(squirrel.Eq{colOwnerID: uid}).ToSql()
	if err != nil {
		return nil, err
	}

	var out user.Groups
	err = r.db.SelectContext(ctx, &out, q, args...)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return out, err
}

// GetGroupMembers implements service.GroupManager
func (r GroupRepository) GetGroupMembers(ctx context.Context, gid user.GroupID) (user.Users, error) {
	q, args, err := psql.Select(colID, colName, colEmail).
		From(tableGroupMembers).InnerJoin(fmt.Sprintf("%s ON %s = %s", tableUsers, colID, colMemberID)).
		Where(squirrel.Eq{colGroupID: gid}).ToSql()

	var out user.Users
	err = r.db.SelectContext(ctx, &out, q, args...)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return out, nil
}
