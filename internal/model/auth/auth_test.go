package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/x1unix/sbda-ledger/internal/model"
)

func TestValidate_LoginRequest(t *testing.T) {
	cases := map[string]struct {
		req     Credentials
		wantErr string
	}{
		"empty": {
			wantErr: "Key: 'Credentials.email' Error:Field validation for 'email' failed on the 'required' tag\n" +
				"Key: 'Credentials.password' Error:Field validation for 'password' failed on the 'required' tag",
		},
		"no email": {
			req:     Credentials{Password: "foobar"},
			wantErr: "Key: 'Credentials.email' Error:Field validation for 'email' failed on the 'required' tag",
		},
		"invalid email": {
			req:     Credentials{Email: "💩", Password: "foobar"},
			wantErr: "Key: 'Credentials.email' Error:Field validation for 'email' failed on the 'email' tag",
		},
		"no password": {
			req:     Credentials{Email: "user@mail.com", Password: ""},
			wantErr: "Key: 'Credentials.password' Error:Field validation for 'password' failed on the 'required' tag",
		},
		"small password": {
			req:     Credentials{Email: "user@mail.com", Password: "a"},
			wantErr: "Key: 'Credentials.password' Error:Field validation for 'password' failed on the 'min' tag",
		},
		"valid struct": {
			req: Credentials{Email: "user@mail.com", Password: "1234567"},
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			checkValidatorErr(t, v.req, v.wantErr)
		})
	}
}

func checkValidatorErr(t *testing.T, val interface{}, expectMsg string) {
	err := model.Validate(val)
	if expectMsg == "" {
		require.NoErrorf(t, err, "unexpected error on validating %#v", val)
		return
	}

	require.EqualError(t, err, expectMsg, "invalid error on validating %#v", val)
}
