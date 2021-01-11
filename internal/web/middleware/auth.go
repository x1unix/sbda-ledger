package middleware

import (
	"net/http"

	"github.com/x1unix/sbda-ledger/internal/model/auth"
	"github.com/x1unix/sbda-ledger/internal/service"
	"github.com/x1unix/sbda-ledger/internal/web"
	"github.com/x1unix/sbda-ledger/internal/web/httputil"
)

// NewAuthMiddleware returns a new middleware which checks if user is authenticated.
//
// If user is authenticated, user session will be populated into request context.
func NewAuthMiddleware(authSvc *service.AuthService) web.MiddlewareFunc {
	return func(rw http.ResponseWriter, req *http.Request) (*http.Request, error) {
		token, ok := httputil.BearerTokenFromRequest(req)
		if !ok {
			return req, service.ErrAuthRequired
		}

		ssid, err := auth.ParseToken(token)
		if err != nil {
			return req, service.ErrAuthRequired
		}

		sess, err := authSvc.GetSession(req.Context(), ssid)
		if err != nil {
			return req, err
		}

		ctx := auth.ContextWithSession(req.Context(), sess)
		return req.WithContext(ctx), nil
	}
}
