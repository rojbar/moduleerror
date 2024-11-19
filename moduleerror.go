// Package moduleerror implements errors for modules usage
package moduleerror

import "fmt"

// Code represents the error code
type Code string

const (
	// CodeInvalidInputError an invalid input has been provided
	CodeInvalidInputError Code = "INVALID_INPUT"

	// CodeNotFoundError resource not found error
	CodeNotFoundError Code = "NOT_FOUND"

	// CodeInternalFailureError internal failure error
	CodeInternalFailureError Code = "INTERNAL_FAILURE"
)

// Error defines the module error interface for all errors used
type Error interface {
	Code() Code
	error
}

type moduleError struct {
	ModuleCode Code `json:"code"`
	err        error
	Message    string `json:"message"`
}

// Code returns the error code
func (e *moduleError) Code() Code {
	return e.ModuleCode
}

// Error returns the error message
func (e *moduleError) Error() string {
	return e.Message
}

// Unwrap unwraps an error
func (e *moduleError) Unwrap() error {
	return e.err
}

// NewInvalidInputError returns an error with CodeInvalidInputError
func NewInvalidInputError(err error) error {
	return &moduleError{
		ModuleCode: CodeInvalidInputError,
		err:        err,
		Message:    fmt.Sprintf("Invalid input provided: %s.", err.Error()),
	}
}

// NewNotFoundError returns an error with CodeNotFoundError
func NewNotFoundError(err error) error {
	return &moduleError{
		ModuleCode: CodeNotFoundError,
		err:        err,
		Message:    fmt.Sprintf("Resource not found: %s.", err.Error()),
	}
}

// NewInternalFailureError returns an error with CodeInternalFailureError
func NewInternalFailureError(err error) error {
	return &moduleError{
		ModuleCode: CodeInternalFailureError,
		err:        err,
		Message:    "An unexpected error occurred.",
	}
}
