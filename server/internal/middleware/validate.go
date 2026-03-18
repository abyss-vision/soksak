package middleware

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	validatorOnce     sync.Once
	globalValidator   *validator.Validate
)

// getValidator returns the singleton validator instance.
func getValidator() *validator.Validate {
	validatorOnce.Do(func() {
		globalValidator = validator.New()
	})
	return globalValidator
}

type fieldError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

// BindAndValidate decodes the JSON request body into T and runs struct validation.
// On a decode or validation error it returns an AppError with status 422.
func BindAndValidate[T any](r *http.Request) (*T, error) {
	var target T

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&target); err != nil {
		return nil, &AppError{
			Code:    "VALIDATION_ERROR",
			Status:  http.StatusUnprocessableEntity,
			Message: "Invalid request body: " + err.Error(),
		}
	}

	if err := getValidator().Struct(target); err != nil {
		var errs validator.ValidationErrors
		if ok := isValidationErrors(err, &errs); ok {
			details := make([]fieldError, 0, len(errs))
			for _, fe := range errs {
				details = append(details, fieldError{
					Field:   fe.Field(),
					Tag:     fe.Tag(),
					Message: fe.Error(),
				})
			}
			return nil, &AppError{
				Code:    "VALIDATION_ERROR",
				Status:  http.StatusUnprocessableEntity,
				Message: "Validation failed",
				Details: details,
			}
		}
		return nil, &AppError{
			Code:    "VALIDATION_ERROR",
			Status:  http.StatusUnprocessableEntity,
			Message: err.Error(),
		}
	}

	return &target, nil
}

func isValidationErrors(err error, out *validator.ValidationErrors) bool {
	if ve, ok := err.(validator.ValidationErrors); ok {
		*out = ve
		return true
	}
	return false
}
