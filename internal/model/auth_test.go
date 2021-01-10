package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate_LoginRequest(t *testing.T) {
	cases := map[string]struct {
		req     LoginRequest
		wantErr string
	}{
		"empty": {
			wantErr: "Key: 'LoginRequest.email' Error:Field validation for 'email' failed on the 'required' tag\n" +
				"Key: 'LoginRequest.password' Error:Field validation for 'password' failed on the 'required' tag",
		},
		"no email": {
			req:     LoginRequest{Password: "foobar"},
			wantErr: "Key: 'LoginRequest.email' Error:Field validation for 'email' failed on the 'required' tag",
		},
		"invalid email": {
			req:     LoginRequest{Email: "💩", Password: "foobar"},
			wantErr: "Key: 'LoginRequest.email' Error:Field validation for 'email' failed on the 'email' tag",
		},
		"no password": {
			req:     LoginRequest{Email: "user@mail.com", Password: ""},
			wantErr: "Key: 'LoginRequest.password' Error:Field validation for 'password' failed on the 'required' tag",
		},
		"small password": {
			req:     LoginRequest{Email: "user@mail.com", Password: "a"},
			wantErr: "Key: 'LoginRequest.password' Error:Field validation for 'password' failed on the 'min' tag",
		},
		"valid struct": {
			req: LoginRequest{Email: "user@mail.com", Password: "1234567"},
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			checkValidatorErr(t, v.req, v.wantErr)
		})
	}
}

func checkValidatorErr(t *testing.T, val interface{}, expectMsg string) {
	err := Validate(val)
	if expectMsg == "" {
		require.NoErrorf(t, err, "unexpected error on validating %#v", val)
		return
	}

	require.EqualError(t, err, expectMsg, "invalid error on validating %#v", val)
}
