package model

import (
	"github.com/x1unix/sbda-ledger/internal/web"
	"net/http"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	// Validator is preconfigured validator instance.
	Validator = validator.New()

	nameRegEx = regexp.MustCompile(`(?m)^[\w ]+$`)
)

func init() {
	must(Validator.RegisterValidation("name", nameValidator))
}

type validatorErrors struct {
	validator.ValidationErrors
}

// APIError implements web.APIErrorer
func (err validatorErrors) APIError() web.APIError {
	errs := make([]validationError, 0, len(err.ValidationErrors))
	for _, err := range err.ValidationErrors {
		errs = append(errs, validationError{
			Namespace: err.Namespace(),
			Field:     err.Field(),
			Validator: err.Tag(),
			Type:      err.Type().String(),
			Param:     err.Param(),
		})
	}

	return web.APIError{
		Status:  http.StatusBadRequest,
		Message: err.ValidationErrors.Error(),
		Data:    errs,
	}
}

func nameValidator(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	return nameRegEx.MatchString(val)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type validationError struct {
	Namespace string `json:"namespace"`
	Field     string `json:"field"`
	Validator string `json:"validator"`
	Type      string `json:"type"`
	Param     string `json:"param"`
}

// Validate performs struct validation and returns an error on failure.
//
// Wraps Validator.Struct method and returns API-compatible error.
func Validate(v interface{}) error {
	err := Validator.Struct(v)
	if err == nil {
		return nil
	}

	valErr, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	return validatorErrors{valErr}
}
