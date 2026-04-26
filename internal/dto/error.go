package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type AppError struct {
	StatusCode int
	Code       string
	Message    string
}

func NewAppError(statusCode int, code string, message string) *AppError {
	return &AppError{StatusCode: statusCode, Code: code, Message: message}
}

func (e *AppError) Error() string {
	return e.Message
}

func ValidationError(err error) error {
	var validationErrors validator.ValidationErrors
	if ok := AsValidationErrors(err, &validationErrors); ok && len(validationErrors) > 0 {
		fieldError := validationErrors[0]
		return NewAppError(422, "VALIDATION_ERROR", fmt.Sprintf("%s is invalid.", fieldError.Field()))
	}
	return NewAppError(422, "VALIDATION_ERROR", "Invalid request.")
}

func AsValidationErrors(err error, target *validator.ValidationErrors) bool {
	value, ok := err.(validator.ValidationErrors)
	if !ok {
		return false
	}
	*target = value
	return true
}
