package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/x1unix/sbda-ledger/internal/model/loan"
	"github.com/x1unix/sbda-ledger/internal/model/user"
	"github.com/x1unix/sbda-ledger/internal/web"
	"go.uber.org/zap"
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

	// GroupMemberIDs returns list of users which are members of group, including group owner.
	GroupMemberIDs(ctx context.Context, gid user.GroupID) ([]user.ID, error)
}

type LoanAdder interface {
	// AddLoan adds a loan for each debtor from lender by specified amount.
	AddLoan(ctx context.Context, lender user.ID, amount loan.Amount, debtors []user.ID) error
}

type GroupService struct {
	log       *zap.Logger
	groups    GroupManager
	loanAdder LoanAdder
}

// NewGroupService is GroupService constructor
func NewGroupService(log *zap.Logger, groups GroupManager, loanAdder LoanAdder) *GroupService {
	return &GroupService{log: log.Named("service.groups"), groups: groups, loanAdder: loanAdder}
}

func (svc GroupService) checkGroupActor(ctx context.Context, actor user.ID, gid user.GroupID) error {
	owner, err := svc.groups.GetGroupOwner(ctx, gid)
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
func (svc GroupService) AddGroup(ctx context.Context, name string, owner user.ID) (*user.Group, error) {
	gid, err := svc.groups.AddGroup(ctx, name, owner)
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
func (svc GroupService) AddMembers(ctx context.Context, actorId user.ID, gid user.GroupID, uids []user.ID) error {
	if len(uids) == 0 {
		return web.NewErrBadRequest("group member list is empty")
	}

	if err := svc.checkGroupActor(ctx, actorId, gid); err != nil {
		return err
	}

	for _, uid := range uids {
		if uid == actorId {
			return web.NewErrBadRequest("group author is already in group")
		}
	}

	return svc.groups.AddGroupUsers(ctx, gid, uids)
}

// GetMembers returns list of group members
func (svc GroupService) GetMembers(ctx context.Context, gid user.GroupID) (user.Users, error) {
	return svc.groups.GetGroupMembers(ctx, gid)
}

// RemoveMember removes a member from a group
func (svc GroupService) RemoveMember(ctx context.Context, actorId user.ID, gid user.GroupID, uid user.ID) error {
	if err := svc.checkGroupActor(ctx, actorId, gid); err != nil {
		return err
	}

	if actorId == uid {
		return web.NewErrBadRequest("group creator cannot be removed from the group")
	}

	return svc.groups.DeleteGroupUser(ctx, gid, uid)
}

// DeleteGroup removes group
func (svc GroupService) DeleteGroup(ctx context.Context, actorId user.ID, gid user.GroupID) error {
	if err := svc.checkGroupActor(ctx, actorId, gid); err != nil {
		return err
	}

	return svc.groups.DeleteGroup(ctx, gid)
}

// GroupsByUser returns all groups where user is owner or member.
func (svc GroupService) GroupsByUser(ctx context.Context, uid user.ID) (user.Groups, error) {
	return svc.groups.GroupsByUser(ctx, uid)
}

// GetGroupInfo returns full group information
func (svc GroupService) GetGroupInfo(ctx context.Context, gid user.GroupID) (*user.GroupInfo, error) {
	g, err := svc.groups.GroupByID(ctx, gid)
	if err == ErrGroupNotFound {
		return nil, web.NewErrNotFound("group not found")
	}

	if err != nil {
		return nil, err
	}

	members, err := svc.groups.GetGroupMembers(ctx, gid)
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}

	return &user.GroupInfo{
		Group:   *g,
		Members: members,
	}, nil
}

func (svc GroupService) ShareExpense(ctx context.Context, actorID user.ID, amount loan.Amount, gid user.GroupID) error {
	members, err := svc.groups.GroupMemberIDs(ctx, gid)
	if err == ErrGroupNotFound {
		return web.NewErrNotFound("group not exists")
	}
	if err != nil {
		return fmt.Errorf("failed to get group member list: %w", err)
	}

	// group should have at least 2 members: owner and guest
	if len(members) < 2 {
		return web.NewErrBadRequest("group is empty")
	}

	// Select everyone except actor ID as debtor,
	// simultaneously check if actor is part of group.
	isActorMember := false
	debtors := make([]user.ID, 0, len(members))
	for _, uid := range members {
		if uid.Bytes == actorID.Bytes {
			isActorMember = true
			continue
		}

		debtors = append(debtors, uid)
	}

	if !isActorMember {
		return web.NewErrForbidden("user is not a member of the group")
	}

	// Calculate debt per member.
	// All prices are represented in cents, and cent is a quantum value
	// (in simple words - there is no thing like "half of cent", it's not Bitcoin).
	//
	// So final value should be rounded, or we gonna lose some money.
	roundedDebt := math.Round(float64(amount) / float64(len(debtors)))
	debtPerUser := loan.Amount(roundedDebt)

	svc.log.Debug("adding a new loan",
		zap.Any("actor_id", actorID),
		zap.Int64("amount_total", amount),
		zap.Int64("amount_per_user", debtPerUser),
		zap.Any("debtors", debtors))

	return svc.loanAdder.AddLoan(ctx, actorID, loan.Amount(roundedDebt), debtors)
}
