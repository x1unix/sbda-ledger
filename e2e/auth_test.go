package e2e

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/x1unix/sbda-ledger/pkg/ledger"
)

func TestAuth_Register(t *testing.T) {
	cases := []struct {
		label   string
		req     ledger.RegisterRequest
		wantErr string
	}{
		{
			label:   "empty payload",
			wantErr: "400 Bad Request: invalid request payload",
		},
		{
			label:   "invalid email",
			wantErr: "400 Bad Request: invalid request payload",
			req: ledger.RegisterRequest{
				Email:    "--",
				Name:     "john",
				Password: "123456",
			},
		},
		{
			label:   "invalid name",
			wantErr: "400 Bad Request: invalid request payload",
			req: ledger.RegisterRequest{
				Email:    "u111@example.com",
				Name:     "",
				Password: "123456",
			},
		},
		{
			label:   "invalid password",
			wantErr: "400 Bad Request: invalid request payload",
			req: ledger.RegisterRequest{
				Email:    "u111@example.com",
				Name:     "joey",
				Password: "",
			},
		},
		{
			label: "valid creds",
			req: ledger.RegisterRequest{
				Email:    "u111@example.com",
				Name:     "joey",
				Password: "123456",
			},
		},
		{
			label:   "duplicate registration",
			wantErr: "400 Bad Request: record already exists",
			req: ledger.RegisterRequest{
				Email:    "u111@example.com",
				Name:     "marko",
				Password: "123456",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			sess, err := Client.Register(c.req)
			if c.wantErr != "" {
				shouldContainError(t, err, c.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, sess.User.Name, c.req.Name)
			require.Equal(t, sess.User.Email, c.req.Email)
		})
	}
}

func TestAuth_Login(t *testing.T) {
	sess, err := Client.Register(ledger.RegisterRequest{
		Email:    "testauthlogin@mail.com",
		Name:     "testauthlogin",
		Password: "123456",
	})
	require.NoError(t, err, "failed to create a user for test case")

	cases := map[string]struct {
		wantErr  string
		creds    ledger.Credentials
		wantUser ledger.User
	}{
		"empty request": {
			wantErr: "400 Bad Request: invalid request payload",
		},
		"empty email": {
			creds:   ledger.Credentials{Password: "12345"},
			wantErr: "400 Bad Request: invalid request payload",
		},
		"empty password": {
			creds:   ledger.Credentials{Email: "testauthlogin@mail.com"},
			wantErr: "400 Bad Request: invalid request payload",
		},
		"invalid username": {
			creds:   ledger.Credentials{Email: "badusername@mail.com", Password: "123456"},
			wantErr: "400 Bad Request: invalid username or password",
		},
		"invalid password": {
			creds:   ledger.Credentials{Email: "testauthlogin@mail.com", Password: "badpassword"},
			wantErr: "400 Bad Request: invalid username or password",
		},
		"valid login": {
			creds:    ledger.Credentials{Email: "testauthlogin@mail.com", Password: "123456"},
			wantUser: sess.User,
		},
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			rsp, err := Client.Login(c.creds)
			if c.wantErr != "" {
				shouldContainError(t, err, c.wantErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, rsp)
			require.Equal(t, c.wantUser, rsp.User)
		})
	}
}

func TestAuth_Session(t *testing.T) {
	sess, err := Client.Register(ledger.RegisterRequest{
		Email:    "testgetsession@mail.com",
		Name:     "testgetsession",
		Password: "123456",
	})
	require.NoError(t, err, "failed to create a user for test case")

	cases := map[string]struct {
		want        ledger.SessionInfo
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
			want:  sess.Session,
		},
		"expired token": {
			token:   sess.Token,
			wantErr: "401 Unauthorized: authorization required",
			onBeforeRun: func(t *testing.T, token *ledger.Token) {
				sess, err := Client.Login(ledger.Credentials{
					Email:    "testgetsession@mail.com",
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

			got, err := Client.Session(token)
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
