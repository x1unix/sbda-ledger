package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/x1unix/sbda-ledger/pkg/ledger"
)

func TestUser_Register(t *testing.T) {
	cases := map[string]struct {
		req     ledger.RegisterRequest
		wantErr string
	}{
		"empty payload": {
			wantErr: "400 Bad Request: invalid request payload",
		},
		"invalid email": {
			wantErr: "400 Bad Request: invalid request payload",
			req: ledger.RegisterRequest{
				Email:    "--",
				Name:     "john",
				Password: "123456",
			},
		},
		"invalid name": {
			wantErr: "400 Bad Request: invalid request payload",
			req: ledger.RegisterRequest{
				Email:    "u111@example.com",
				Name:     "",
				Password: "123456",
			},
		},
		"invalid password": {
			wantErr: "400 Bad Request: invalid request payload",
			req: ledger.RegisterRequest{
				Email:    "u111@example.com",
				Name:     "joey",
				Password: "",
			},
		},
		"valid creds": {
			req: ledger.RegisterRequest{
				Email:    "u111@example.com",
				Name:     "joey",
				Password: "123456",
			},
		},
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
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
