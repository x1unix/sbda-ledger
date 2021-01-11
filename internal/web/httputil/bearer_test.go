package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBearerTokenFromRequest(t *testing.T) {
	cases := map[string]struct {
		input     string
		expectVal string
		expectOk  bool
	}{
		"valid bearer token": {
			input:     "Bearer foo",
			expectVal: "foo",
			expectOk:  true,
		},
		"valid bearer token with padding": {
			input:     "Bearer    foo ",
			expectVal: "foo",
			expectOk:  true,
		},
		"empty header": {
			input: "",
		},
		"header without value": {
			input: "Bearer ",
		},
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://[:::]", nil)
			req.Header.Set(authHeader, c.expectVal)
			got, ok := BearerTokenFromRequest(req)
			require.Equal(t, c.expectVal, got)
			require.Equal(t, c.expectOk, ok)
		})
	}
}
