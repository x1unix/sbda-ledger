package web

import "net/http"

// APIError is HTTP error returned from API
type APIError struct {
	// Status is HTTP status code
	Status int `json:"-"`

	// Message is error message
	Message string `json:"message"`

	// Data is optional error data
	Data interface{} `json:"data,omitempty"`
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

// NewBadRequestError returns a new bad request API error
func NewBadRequestError(msg string) *APIError {
	return &APIError{Status: http.StatusBadRequest, Message: msg}
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
