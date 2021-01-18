package e2e

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/x1unix/sbda-ledger/pkg/ledger"
)

func TestUser_GetUsersList(t *testing.T) {
	const usersCount = 5
	require.NoError(t, TruncateData(), "failed to truncate data from the test")
	var tkn ledger.Token
	expects := make([]ledger.User, usersCount)
	for i := 0; i < usersCount; i++ {
		sess, err := Client.Register(ledger.RegisterRequest{
			Email:    fmt.Sprintf("getuserslist%d@mail.com", i),
			Name:     fmt.Sprintf("getuserslist%d", i),
			Password: "123456",
		})
		require.NoError(t, err, "failed to create a user for test case")
		if i == 0 {
			tkn = sess.Token
		}

		expects[i] = sess.User
	}

	cases := map[string]struct {
		want        []ledger.User
		wantErr     string
		token       ledger.Token
		onBeforeRun func(t *testing.T, token *ledger.Token)
	}{
		"empty token": {
			wantErr: "401 Unauthorized: authorization required",
		},
		"invalid token": {
			wantErr: "401 Unauthorized: authorization required",
			token:   ledger.Token(uuid.New().String()),
		},
		"valid token": {
			token: tkn,
			want:  expects,
		},
		"expired token": {
			wantErr: "401 Unauthorized: authorization required",
			onBeforeRun: func(t *testing.T, token *ledger.Token) {
				sess, err := Client.Login(ledger.Credentials{
					Email:    "getuserslist1@mail.com",
					Password: "123456",
				})
				require.NoError(t, err, "failed to login as user for test case")
				*token = sess.Token

				require.NoError(t, Client.Logout(sess.Token), "failed to logout from test account")
			},
		},
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			token := c.token
			if c.onBeforeRun != nil {
				c.onBeforeRun(t, &token)
			}

			got, err := Client.Users(token)
			if c.wantErr != "" {
				shouldContainError(t, err, c.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, c.want, got)
		})
	}
}

func TestUser_Current(t *testing.T) {
	sess, err := Client.Register(ledger.RegisterRequest{
		Email:    "testgetcurrentuser@mail.com",
		Name:     "testgetcurrentuser",
		Password: "123456",
	})
	require.NoError(t, err, "failed to create a user for test case")

	cases := map[string]struct {
		want        ledger.User
		wantErr     string
		token       ledger.Token
		onBeforeRun func(t *testing.T, token *ledger.Token)
	}{
		"empty token": {
			wantErr: "401 Unauthorized: authorization required",
		},
		"invalid token": {
			wantErr: "401 Unauthorized: authorization required",
			token:   ledger.Token(uuid.New().String()),
		},
		"valid token": {
			token: sess.Token,
			want:  sess.User,
		},
		"expired token": {
			token:   sess.Token,
			wantErr: "401 Unauthorized: authorization required",
			onBeforeRun: func(t *testing.T, token *ledger.Token) {
				sess, err := Client.Login(ledger.Credentials{
					Email:    "testgetcurrentuser@mail.com",
					Password: "123456",
				})
				require.NoError(t, err, "failed to create a user for test case")
				*token = sess.Token

				require.NoError(t, Client.Logout(sess.Token), "failed to logout from test account")
			},
		},
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			token := c.token
			if c.onBeforeRun != nil {
				c.onBeforeRun(t, &token)
			}

			got, err := Client.CurrentUser(token)
			if c.wantErr != "" {
				shouldContainError(t, err, c.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, c.want, *got)
		})
	}
}

func TestUser_GetByID(t *testing.T) {
	sess, err := Client.Register(ledger.RegisterRequest{
		Email:    "testuserbyid1@mail.com",
		Name:     "testuserbyid1",
		Password: "123456",
	})
	require.NoError(t, err, "failed to create a user for test case")

	sess2, err := Client.Register(ledger.RegisterRequest{
		Email:    "testuserbyid2@mail.com",
		Name:     "testuserbyid2",
		Password: "123456",
	})
	require.NoError(t, err, "failed to create a user for test case")

	cases := map[string]struct {
		want        ledger.User
		wantErr     string
		token       ledger.Token
		uid         string
		onBeforeRun func(t *testing.T, token *ledger.Token)
	}{
		"empty token": {
			wantErr: "401 Unauthorized: authorization required",
			uid:     sess2.User.ID,
		},
		"invalid token": {
			wantErr: "401 Unauthorized: authorization required",
			uid:     sess2.User.ID,
			token:   ledger.Token(uuid.New().String()),
		},
		"expired token": {
			token:   sess2.Token,
			uid:     sess2.User.ID,
			wantErr: "401 Unauthorized: authorization required",
			onBeforeRun: func(t *testing.T, token *ledger.Token) {
				require.NoError(t, Client.Logout(*token), "failed to logout from test account")
			},
		},
		"no such user": {
			token:   sess.Token,
			uid:     uuid.New().String(),
			wantErr: "404 Not Found: user not found",
		},
		"valid user": {
			token: sess.Token,
			uid:   sess2.User.ID,
			want:  sess2.User,
		},
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			token := c.token
			if c.onBeforeRun != nil {
				c.onBeforeRun(t, &token)
			}

			got, err := Client.UserByID(c.uid, token)
			if c.wantErr != "" {
				shouldContainError(t, err, c.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, c.want, *got)
		})
	}
}
