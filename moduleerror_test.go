package moduleerror_test

import (
	"errors"
	"testing"

	"github.com/rojbar/moduleerror"
)

func TestNewInvalidInputError(t *testing.T) {
	dummyError := errors.New("dummy")

	err := moduleerror.NewInvalidInputError(dummyError)
	if !errors.Is(err, dummyError) {
		t.Errorf("err is not dummyError %v %v", err, dummyError)
	}

	var moduleError moduleerror.Error

	if !errors.As(err, &moduleError) {
		t.Errorf("err not as moduleerror.Error %v %v", err, dummyError)
	}

	if moduleError.Code() != moduleerror.CodeInvalidInputError {
		t.Errorf("expected moduleError to have Code value %s instead got %s", moduleerror.CodeInvalidInputError, moduleError.Code())

	}
	if moduleError.Error() != "Invalid input provided: dummy." {
		t.Errorf("expected moduleError to have message value %s instead got %s", "Invalid input provided: dummy.", moduleError.Error())
	}
}

func TestNewNotFoundError(t *testing.T) {
	dummyError := errors.New("dummy")

	err := moduleerror.NewNotFoundError(dummyError)
	if !errors.Is(err, dummyError) {
		t.Errorf("err is not dummyError %v %v", err, dummyError)
	}

	var moduleError moduleerror.Error

	if !errors.As(err, &moduleError) {
		t.Errorf("err not as moduleerror.Error %v %v", err, dummyError)
	}

	if moduleError.Code() != moduleerror.CodeNotFoundError {
		t.Errorf("expected moduleError to have Code value %s instead got %s", moduleerror.CodeNotFoundError, moduleError.Code())
	}

	if moduleError.Error() != "Resource not found: dummy." {
		t.Errorf("expected moduleError to have message value %s instead got %s", "Resource not found: dummy.", moduleError.Error())
	}
}

func TestInternalFailureError(t *testing.T) {
	dummyError := errors.New("dummy")

	err := moduleerror.NewInternalFailureError(dummyError)
	if !errors.Is(err, dummyError) {
		t.Errorf("err is not dummyError %v %v", err, dummyError)
	}

	var moduleError moduleerror.Error

	if !errors.As(err, &moduleError) {
		t.Errorf("err not as moduleerror.Error %v %v", err, dummyError)
	}

	if moduleError.Code() != moduleerror.CodeInternalFailureError {
		t.Errorf("expected moduleError to have Code value %s instead got %s", moduleerror.CodeInternalFailureError, moduleError.Code())
	}

	if moduleError.Error() != "An unexpected error occurred." {
		t.Errorf("expected moduleError to have message value %s instead got %s", "An unexpected error occurred.", moduleError.Error())
	}
}
