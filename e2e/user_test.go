package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/x1unix/sbda-ledger/pkg/ledger"
)

func TestUser_Register(t *testing.T) {
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
