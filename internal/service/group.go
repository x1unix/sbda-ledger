package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/x1unix/sbda-ledger/internal/model/user"
	"github.com/x1unix/sbda-ledger/internal/web"
)

var (
	ErrGroupNotFound = errors.New("group not found")
)

// GroupStore stores group
type GroupStore interface {
	AddGroup(ctx context.Context, name string, owner user.ID) (*user.GroupID, error)
	DeleteGroup(ctx context.Context, gid user.GroupID) error
	GroupByID(ctx context.Context, gid user.ID) (*user.Group, error)
	GetGroupOwner(ctx context.Context, gid user.GroupID) (*user.ID, error)
}

// GroupManager manages group information and members list
type GroupManager interface {
	GroupStore

	// AddGroupUsers adds a new user to a group
	AddGroupUsers(ctx context.Context, gid user.GroupID, uids []user.ID) error

	// DeleteGroupUser removes user from group
	DeleteGroupUser(ctx context.Context, gid user.GroupID, uid user.ID) error

	// GetGroupMembers returns group members (except owner)
	GetGroupMembers(ctx context.Context, gid user.GroupID) (user.Users, error)

	// GroupsByUser returns all groups where user is owner or member.
	GroupsByUser(ctx context.Context, uid user.ID) (user.Groups, error)
}

type GroupService struct {
	groups GroupManager
}

// NewGroupService is GroupService constructor
func NewGroupService(groups GroupManager) *GroupService {
	return &GroupService{groups: groups}
}

func (r GroupService) checkGroupActor(ctx context.Context, actor user.ID, gid user.GroupID) error {
	owner, err := r.groups.GetGroupOwner(ctx, gid)
	if err == ErrGroupNotFound {
		return web.NewErrNotFound("group not found")
	}

	if err != nil {
		return err
	}

	if owner == nil {
		return errors.New("bad group owner")
	}

	if *owner != actor {
		return web.NewErrForbidden("you have no right to control this group")
	}

	return nil
}

// AddGroup creates a new group
func (r GroupService) AddGroup(ctx context.Context, name string, owner user.ID) (*user.Group, error) {
	gid, err := r.groups.AddGroup(ctx, name, owner)
	if err != nil {
		return nil, err
	}

	return &user.Group{
		ID:      *gid,
		Name:    name,
		OwnerID: owner,
	}, nil
}

// AddMembers adds members to a group
func (r GroupService) AddMembers(ctx context.Context, actorId user.ID, gid user.GroupID, uids []user.ID) error {
	if len(uids) == 0 {
		return web.NewErrBadRequest("group member list is empty")
	}

	if err := r.checkGroupActor(ctx, actorId, gid); err != nil {
		return err
	}

	return r.groups.AddGroupUsers(ctx, gid, uids)
}

// GetMembers returns list of group members
func (r GroupService) GetMembers(ctx context.Context, gid user.GroupID) (user.Users, error) {
	return r.groups.GetGroupMembers(ctx, gid)
}

// RemoveMember removes a member from a group
func (r GroupService) RemoveMember(ctx context.Context, actorId user.ID, gid user.GroupID, uid user.ID) error {
	if err := r.checkGroupActor(ctx, actorId, gid); err != nil {
		return err
	}

	return r.groups.DeleteGroupUser(ctx, gid, uid)
}

// DeleteGroup removes group
func (r GroupService) DeleteGroup(ctx context.Context, actorId user.ID, gid user.GroupID) error {
	if err := r.checkGroupActor(ctx, actorId, gid); err != nil {
		return err
	}

	return r.groups.DeleteGroup(ctx, gid)
}

// GroupsByUser returns all groups where user is owner or member.
func (r GroupService) GroupsByUser(ctx context.Context, uid user.ID) (user.Groups, error) {
	return r.groups.GroupsByUser(ctx, uid)
}

// GetGroupInfo returns full group information
func (r GroupService) GetGroupInfo(ctx context.Context, gid user.GroupID) (*user.GroupInfo, error) {
	g, err := r.groups.GroupByID(ctx, gid)
	if err == ErrGroupNotFound {
		return nil, web.NewErrNotFound("group not found")
	}

	if err != nil {
		return nil, err
	}

	members, err := r.groups.GetGroupMembers(ctx, gid)
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}

	return &user.GroupInfo{
		Group:   *g,
		Members: members,
	}, nil
}
