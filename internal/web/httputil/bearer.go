package httputil

import (
	"net/http"
	"strings"
)

const (
	authHeader   = "Authorization"
	bearerPrefix = "Bearer "
)

// BearerTokenFromRequest returns bearer token from request
func BearerTokenFromRequest(r *http.Request) (token string, ok bool) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", false
	}

	bearer := strings.TrimSpace(strings.TrimPrefix(auth, bearerPrefix))
	return bearer, bearer != ""
}
