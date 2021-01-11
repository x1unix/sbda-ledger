package web

import (
	"fmt"
	"net/http"
)

var ErrNotImplemented = NewAPIError(http.StatusNotImplemented, "method not implemented")

// APIError is HTTP error returned from API
type APIError struct {
	// Status is HTTP status code
	Status int `json:"-"`

	// Message is error message
	Message string `json:"message"`

	// Data is optional error data
	Data interface{} `json:"data,omitempty"`
}

// NewAPIError constructs new API error
func NewAPIError(status int, format string, args ...interface{}) *APIError {
	if len(args) > 0 {
		format = fmt.Sprintf(format, args...)
	}

	return &APIError{
		Status:  status,
		Message: format,
	}
}

// Error implements error interface
func (err APIError) Error() string {
	return err.Message
}

// APIErrorer provides and APIError representation of error.
//
// Can be used to implement custom error response.
type APIErrorer interface {
	// APIError returns api error response
	APIError() *APIError
}

// NewErrBadRequest returns a new bad request API error
func NewErrBadRequest(msg string, args ...interface{}) *APIError {
	return NewAPIError(http.StatusBadRequest, msg, args...)
}

// NewErrUnauthorized returns a new unauthorized API error
func NewErrUnauthorized(msg string, args ...interface{}) *APIError {
	return NewAPIError(http.StatusUnauthorized, msg, args...)
}

// NewErrNotFound returns new not found error
func NewErrNotFound(msg string, args ...interface{}) *APIError {
	return NewAPIError(http.StatusNotFound, msg, args...)
}

// NewErrForbidden returns new forbidden error
func NewErrForbidden(msg string, args ...interface{}) *APIError {
	return NewAPIError(http.StatusForbidden, msg, args...)
}

// ToAPIError constructs APIError from passed error.
//
// If error implements APIErrorer interface, APIError() method will be called.
func ToAPIError(err error) *APIError {
	switch t := err.(type) {
	case APIErrorer:
		return t.APIError()
	case *APIError:
		return t
	default:
		return &APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
}
