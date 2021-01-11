package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// ErrorResponse is server error response
type ErrorResponse struct {
	// Error contains server error
	Error *APIError `json:"error"`
}

// HandlerFunc is http.HandlerFunc extension which can return error.
type HandlerFunc = func(rw http.ResponseWriter, req *http.Request) error

// ResourceHandlerFunc is http.HandlerFunc extension which can return
// result which will be encoded to JSON or response error.
type ResourceHandlerFunc = func(req *http.Request) (interface{}, error)

// MiddlewareFunc is request wrapper
type MiddlewareFunc = func(rw http.ResponseWriter, req *http.Request) (*http.Request, error)

// Wrapper is http handler wrapper and composer.
type Wrapper struct {
	log *zap.Logger
}

// NewWrapper is Wrapper constructor
func NewWrapper(log *zap.Logger) *Wrapper {
	return &Wrapper{log: log}
}

// WrapHandler wraps web's HandlerFunc onto http.HandlerFunc.
//
// Accepts optional list of middleware functions to be called before handler.
//
// Examples:
//
//	// one handler
//	web.WrapHandler(myHandler)
//
//	// handler with multiple middlewares
//	web.WrapHandler(getUserData, RequireCORS, RequireAuth)
//
func (w Wrapper) WrapHandler(handler HandlerFunc, wrappers ...MiddlewareFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				w.serveResponseError(rw, fmt.Errorf("panic occured: %v", r))
			}
		}()

		if len(wrappers) == 0 {
			w.serveResponseError(rw, handler(rw, r))
			return
		}

		var err error
		for _, mw := range wrappers {
			r, err = mw(rw, r)
			if err != nil {
				w.serveResponseError(rw, err)
				return
			}
		}

		w.serveResponseError(rw, handler(rw, r))
	}
}

// WrapResourceHandler wraps resource handler onto http.HandlerFunc.
// Use *web.APIError or implement web.APIErrorer to return custom error.
//
// Accepts optional list of middleware functions to be called before handler.
//
// See: WrapHandler
func (w Wrapper) WrapResourceHandler(h ResourceHandlerFunc, mw ...MiddlewareFunc) http.HandlerFunc {
	return w.WrapHandler(func(rw http.ResponseWriter, req *http.Request) error {
		rw.Header().Set("Content-Type", "application/json")
		obj, err := h(req)
		if err != nil {
			return err
		}

		data, err := json.Marshal(obj)
		if err != nil {
			return fmt.Errorf("failed to encode response: %w", err)
		}

		if _, err = rw.Write(data); err != nil {
			// request connection is corrupted, just log error and exit
			w.log.Error("failed to serve response", zap.Error(err))
		}

		return nil
	}, mw...)
}

// MiddlewareFunc wraps web's middleware onto mux.MiddlewareFunc.
func (w Wrapper) MiddlewareFunc(fn MiddlewareFunc) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			newReq, err := fn(rw, r)
			if err != nil {
				w.serveResponseError(rw, err)
				return
			}

			next.ServeHTTP(rw, newReq)
		})
	}
}

func (w Wrapper) serveResponseError(rw http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	apiErr := ToAPIError(err)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(apiErr.Status)
	if apiErr.Status >= http.StatusInternalServerError {
		// Log critical response errors
		w.log.Error(err.Error(), zap.Int("status", apiErr.Status))
	}

	resp := ErrorResponse{Error: apiErr}
	if err := json.NewEncoder(rw).Encode(resp); err != nil {
		w.log.Error("failed to encode error response", zap.Error(err))
	}
}
