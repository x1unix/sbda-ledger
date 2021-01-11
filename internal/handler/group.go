package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/x1unix/sbda-ledger/internal/model"
	"github.com/x1unix/sbda-ledger/internal/model/auth"
	"github.com/x1unix/sbda-ledger/internal/model/request"
	"github.com/x1unix/sbda-ledger/internal/service"
)

type GroupHandler struct {
	groupService *service.GroupService
}

// NewGroupHandler is GroupHandler constructor
func NewGroupHandler(groupSvc *service.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupSvc}
}

func (h GroupHandler) CreateGroup(r *http.Request) (interface{}, error) {
	ctx := r.Context()
	sess := auth.SessionFromContext(ctx)
	if sess == nil {
		return nil, service.ErrAuthRequired
	}

	var req request.GroupCreateRequest
	if err := UnmarshalAndValidate(r.Body, &r); err != nil {
		return nil, err
	}

	return h.groupService.AddGroup(ctx, req.Name, sess.UserID)
}

func (h GroupHandler) GetGroupInfo(r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	gid, err := model.DecodeUUID(vars["groupId"])
	if err != nil {
		return nil, err
	}

	return h.groupService.GetGroupInfo(r.Context(), *gid)
}

func (h GroupHandler) GetUserGroups(r *http.Request) (interface{}, error) {
	ctx := r.Context()
	sess := auth.SessionFromContext(ctx)
	if sess == nil {
		return nil, service.ErrAuthRequired
	}

	groups, err := h.groupService.GroupsByUser(ctx, sess.UserID)
	if err != nil {
		return nil, err
	}

	return request.GroupsResponse{Groups: groups}, nil
}

func (h GroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	sess := auth.SessionFromContext(ctx)
	if sess == nil {
		return service.ErrAuthRequired
	}

	vars := mux.Vars(r)
	gid, err := model.DecodeUUID(vars["groupId"])
	if err != nil {
		return err
	}

	err = h.groupService.DeleteGroup(ctx, sess.UserID, *gid)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h GroupHandler) GetMembers(r *http.Request) (interface{}, error) {
	ctx := r.Context()
	sess := auth.SessionFromContext(ctx)
	if sess == nil {
		return nil, service.ErrAuthRequired
	}

	vars := mux.Vars(r)
	gid, err := model.DecodeUUID(vars["groupId"])
	if err != nil {
		return nil, err
	}

	members, err := h.groupService.GetMembers(ctx, *gid)
	if err != nil {
		return nil, err
	}

	return request.UsersList{Users: members}, nil
}

func (h GroupHandler) AddMembers(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	sess := auth.SessionFromContext(ctx)
	if sess == nil {
		return service.ErrAuthRequired
	}

	vars := mux.Vars(r)
	gid, err := model.DecodeUUID(vars["groupId"])
	if err != nil {
		return err
	}

	var req request.UserIDs
	if err = UnmarshalAndValidate(r.Body, &req); err != nil {
		return err
	}

	err = h.groupService.AddMembers(ctx, sess.UserID, *gid, req.IDs)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h GroupHandler) RemoveMember(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	sess := auth.SessionFromContext(ctx)
	if sess == nil {
		return service.ErrAuthRequired
	}

	vars := mux.Vars(r)
	ids, err := model.DecodeUUIDs(vars["groupId"], vars["userId"])
	if err != nil {
		return err
	}

	err = h.groupService.RemoveMember(ctx, sess.UserID, ids[0], ids[1])
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
