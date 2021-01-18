package e2e

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/x1unix/sbda-ledger/pkg/ledger"
)

func TestGroup_Create(t *testing.T) {
	groupOwner := mustCreateUser(t, "testgroupcreate", "testgroupcreate@mail.com")

	cases := []struct {
		label     string
		wantErr   string
		groupName string
		token     ledger.Token
		expect    ledger.GroupInfo
	}{
		{
			label:     "require auth",
			groupName: "foobar",
			wantErr:   "401 Unauthorized: authorization required",
		},
		{
			label:     "require token",
			groupName: "foobar",
			token:     ledger.Token(uuid.New().String()),
			wantErr:   "401 Unauthorized: authorization required",
		},
		{
			label:     "invalid name",
			groupName: "",
			token:     groupOwner.Token,
			wantErr:   "400 Bad Request: invalid request payload",
		},
		{
			label:     "valid group",
			groupName: "testgroupcreate",
			token:     groupOwner.Token,
			expect: ledger.GroupInfo{
				Group: ledger.Group{
					OwnerID: groupOwner.User.ID,
					Name:    groupOwner.User.Name,
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			rsp, err := Client.CreateGroup(c.groupName, c.token)
			if c.wantErr != "" {
				require.Error(t, err)
				shouldContainError(t, err, c.wantErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, rsp)
			require.Equal(t, rsp.Name, c.expect.Name)
			require.Equal(t, rsp.OwnerID, c.expect.OwnerID)

			info, err := Client.GroupByID(rsp.ID, c.token)
			require.NoError(t, err)
			require.NotNil(t, info)
			require.Equal(t, rsp.ID, info.ID)
			require.Equal(t, c.expect.Name, info.Name)
			require.Equal(t, c.expect.OwnerID, info.OwnerID)
			require.Equal(t, c.expect.Members, info.Members)
		})
	}
}

func TestGroup_DeleteMembers(t *testing.T) {
	groupOwner := mustCreateUser(t, "testgroupdelmem", "testgroupdelmem@mail.com")
	groupMember := mustCreateUser(t, "testgroupdelmemmem", "testgroupdelmemmem@mail.com")
	grp, err := Client.CreateGroup("testgroupdelmem", groupOwner.Token)
	require.NoError(t, err, "failed to create a test group")
	cases := []struct {
		label     string
		wantErr   string
		groupID   string
		memberID  string
		token     ledger.Token
		beforeRun func(t *testing.T)
		afterRun  func(t *testing.T)
	}{
		{
			label:    "require auth",
			groupID:  uuid.New().String(),
			memberID: uuid.New().String(),
			wantErr:  "401 Unauthorized: authorization required",
		},
		{
			label:    "require token",
			groupID:  uuid.New().String(),
			memberID: uuid.New().String(),
			token:    ledger.Token(uuid.New().String()),
			wantErr:  "401 Unauthorized: authorization required",
		},
		{
			label:    "invalid group id",
			groupID:  "fubar",
			memberID: uuid.New().String(),
			token:    groupOwner.Token,
			wantErr:  "400 Bad Request: invalid resource id: invalid length for UUID",
		},
		{
			label:    "group not exists",
			groupID:  uuid.New().String(),
			token:    groupOwner.Token,
			memberID: uuid.New().String(),
			wantErr:  "404 Not Found: group not found",
		},
		{
			label:    "invalid member id",
			groupID:  uuid.New().String(),
			token:    groupOwner.Token,
			memberID: "xx",
			wantErr:  "400 Bad Request: invalid resource id",
		},
		{
			label:    "other people can't modify group",
			groupID:  grp.ID,
			token:    groupMember.Token,
			memberID: groupMember.User.ID,
			wantErr:  "403 Forbidden: you have no right to control this group",
		},
		{
			label:    "owner cannot remove itself",
			groupID:  grp.ID,
			token:    groupOwner.Token,
			memberID: groupOwner.User.ID,
			wantErr:  "400 Bad Request: group creator cannot be removed from the group",
		},
		{
			label:    "members except owner cant' delete members",
			groupID:  grp.ID,
			token:    groupMember.Token,
			memberID: groupMember.User.ID,
			wantErr:  "403 Forbidden: you have no right to control this group",
			beforeRun: func(t *testing.T) {
				require.NoError(t, Client.AddGroupMembers(grp.ID, groupOwner.Token, groupMember.User.ID),
					"can't add test member to a test group")
			},
			afterRun: func(t *testing.T) {
				require.NoError(t, Client.DeleteGroupMember(grp.ID, groupMember.User.ID, groupOwner.Token),
					"can't remove test member")
			},
		},
		{
			label:    "user not a member",
			groupID:  grp.ID,
			token:    groupOwner.Token,
			memberID: uuid.New().String(),
			wantErr:  "404 Not Found: item not found",
		},
		{
			label:    "valid group",
			groupID:  grp.ID,
			token:    groupOwner.Token,
			memberID: groupMember.User.ID,
			beforeRun: func(t *testing.T) {
				require.NoError(t, Client.AddGroupMembers(grp.ID, groupOwner.Token, groupMember.User.ID),
					"can't add test member to a test group")
			},
		},
	}

	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			if c.beforeRun != nil {
				c.beforeRun(t)
			}
			err := Client.DeleteGroupMember(c.groupID, c.memberID, c.token)
			if c.wantErr != "" {
				require.Error(t, err)
				shouldContainError(t, err, c.wantErr)
				if c.afterRun != nil {
					c.afterRun(t)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGroup_AddMembers(t *testing.T) {
	groupOwner := mustCreateUser(t, "testgroupaddmem", "testgroupaddmem@mail.com")
	groupMember := mustCreateUser(t, "testgroupaddmemmem", "testgroupaddmemmem@mail.com")
	grp, err := Client.CreateGroup("testgroupaddmem", groupOwner.Token)
	require.NoError(t, err, "failed to create a test group")
	cases := []struct {
		label     string
		wantErr   string
		groupID   string
		memberID  string
		token     ledger.Token
		beforeRun func(t *testing.T)
		afterRun  func(t *testing.T)
	}{
		{
			label:    "require auth",
			groupID:  uuid.New().String(),
			memberID: uuid.New().String(),
			wantErr:  "401 Unauthorized: authorization required",
		},
		{
			label:    "require token",
			groupID:  uuid.New().String(),
			memberID: uuid.New().String(),
			token:    ledger.Token(uuid.New().String()),
			wantErr:  "401 Unauthorized: authorization required",
		},
		{
			label:    "invalid group id",
			groupID:  "fubar",
			memberID: uuid.New().String(),
			token:    groupOwner.Token,
			wantErr:  "400 Bad Request: invalid resource id: invalid length for UUID",
		},
		{
			label:    "group not exists",
			groupID:  uuid.New().String(),
			token:    groupOwner.Token,
			memberID: uuid.New().String(),
			wantErr:  "404 Not Found: group not found",
		},
		{
			label:    "invalid member id",
			groupID:  uuid.New().String(),
			token:    groupOwner.Token,
			memberID: "xx",
			wantErr:  "400 Bad Request: invalid resource id",
		},
		{
			label:    "other people can't modify group",
			groupID:  grp.ID,
			token:    groupMember.Token,
			memberID: groupMember.User.ID,
			wantErr:  "403 Forbidden: you have no right to control this group",
		},
		{
			label:    "owner cannot remove itself",
			groupID:  grp.ID,
			token:    groupOwner.Token,
			memberID: groupOwner.User.ID,
			wantErr:  "400 Bad Request: group creator cannot be removed from the group",
		},
		{
			label:    "members except owner cant add members",
			groupID:  grp.ID,
			token:    groupMember.Token,
			memberID: groupMember.User.ID,
			wantErr:  "403 Forbidden: you have no right to control this group",
			beforeRun: func(t *testing.T) {
				require.NoError(t, Client.AddGroupMembers(grp.ID, groupOwner.Token, groupMember.User.ID),
					"can't add test member to a test group")
			},
			afterRun: func(t *testing.T) {
				require.NoError(t, Client.DeleteGroupMember(grp.ID, groupMember.User.ID, groupOwner.Token),
					"can't remove test member")
			},
		},
		{
			label:    "user not a member",
			groupID:  grp.ID,
			token:    groupOwner.Token,
			memberID: uuid.New().String(),
			wantErr:  "404 Not Found: item not found",
		},
		{
			label:    "valid group",
			groupID:  grp.ID,
			token:    groupOwner.Token,
			memberID: groupMember.User.ID,
			beforeRun: func(t *testing.T) {
				require.NoError(t, Client.AddGroupMembers(grp.ID, groupOwner.Token, groupMember.User.ID),
					"can't add test member to a test group")
			},
		},
	}

	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			if c.beforeRun != nil {
				c.beforeRun(t)
			}
			err := Client.DeleteGroupMember(c.groupID, c.memberID, c.token)
			if c.wantErr != "" {
				require.Error(t, err)
				shouldContainError(t, err, c.wantErr)
				if c.afterRun != nil {
					c.afterRun(t)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGroup_Delete(t *testing.T) {
	groupOwner := mustCreateUser(t, "testgroupdelete", "testgroupdelete@mail.com")
	groupMember := mustCreateUser(t, "testgroupdeletemem", "testgroupdeletemem@mail.com")
	grp, err := Client.CreateGroup("testgroupdelete", groupOwner.Token)
	require.NoError(t, err, "failed to create a test group")
	cases := []struct {
		label     string
		wantErr   string
		groupID   string
		token     ledger.Token
		beforeRun func(t *testing.T)
	}{
		{
			label:   "require auth",
			groupID: uuid.New().String(),
			wantErr: "401 Unauthorized: authorization required",
		},
		{
			label:   "require token",
			groupID: uuid.New().String(),
			token:   ledger.Token(uuid.New().String()),
			wantErr: "401 Unauthorized: authorization required",
		},
		{
			label:   "invalid id",
			groupID: "fubar",
			token:   groupOwner.Token,
			wantErr: "400 Bad Request: invalid resource id: invalid length for UUID",
		},
		{
			label:   "id not exists",
			groupID: uuid.New().String(),
			token:   groupOwner.Token,
			wantErr: "404 Not Found: group not found",
		},
		{
			label:   "other guests can't delete group",
			groupID: grp.ID,
			token:   groupMember.Token,
			wantErr: "403 Forbidden: you have no right to control this group",
		},
		{
			label:   "guests can't delete group",
			groupID: grp.ID,
			token:   groupMember.Token,
			wantErr: "403 Forbidden: you have no right to control this group",
			beforeRun: func(t *testing.T) {
				require.NoError(t, Client.AddGroupMembers(grp.ID, groupOwner.Token, groupMember.User.ID),
					"can't add test member to a test group")

			},
		},
		{
			label:   "valid group",
			groupID: grp.ID,
			token:   groupOwner.Token,
		},
	}

	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			if c.beforeRun != nil {
				c.beforeRun(t)
			}
			err := Client.DeleteGroup(c.groupID, c.token)
			if c.wantErr != "" {
				require.Error(t, err)
				shouldContainError(t, err, c.wantErr)
				return
			}

			require.NoError(t, err)
			_, err = Client.GroupByID(c.groupID, c.token)
			require.Error(t, err)
			v, ok := err.(*ledger.ErrorResponse)
			if !ok {
				t.Fatalf("returned inner error is not %T, but %T", v, err)
				return
			}

			require.Equal(t, http.StatusNotFound, v.StatusCode, "unexpected response status from deleted group")
		})
	}
}

//func TestGroup(t *testing.T) {
//	groupOwner := mustCreateUser(t, "testgroupowner", "testgroup0@mail.com")
//	orphanUser := mustCreateUser(t, "testgrouporphan", "testgrouporphan@mail.com")
//	testGroupMember1 := mustCreateUser(t, "testgroupmem1", "testgroup1@mail.com")
//	testGroupMember2 := mustCreateUser(t, "testgroupmem2", "testgroup1@mail.com")
//
//}
