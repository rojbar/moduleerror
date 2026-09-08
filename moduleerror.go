// Package moduleerror implements errors for modules usage.
//
// An error carries a Code that classifies the failure without naming any
// transport, so the same value can become an HTTP status, an MCP tool
// failure, or a gRPC code at whichever boundary handles it.
//
// Each error exposes two texts, for two different audiences:
//
//   - Error returns the code followed by the whole wrapped chain. It is for
//     operators, so logging the value logs everything underneath it.
//   - ClientMessage returns what is safe to hand back to the caller. For an
//     internal failure that is a fixed generic sentence and never the cause.
//
// That split is what lets one service call another. The caller logs the whole
// chain the callee produced, while the client of the outermost service only
// ever sees the generic message.
package moduleerror

// Code represents the error code.
type Code string

const (
	// CodeInvalidInputError an invalid input has been provided.
	CodeInvalidInputError Code = "INVALID_INPUT"

	// CodeNotFoundError resource not found error.
	CodeNotFoundError Code = "NOT_FOUND"

	// CodeInternalFailureError internal failure error.
	CodeInternalFailureError Code = "INTERNAL_FAILURE"

	// CodeConflictError conflict error.
	CodeConflictError Code = "CONFLICT"
)

// internalFailureMessage is the only text an internal failure ever exposes to
// a caller. The cause stays in Error, which is for operators.
const internalFailureMessage = "an unexpected error occurred"

// Error defines the module error interface for all errors used.
type Error interface {
	error

	// Code returns the classification of the failure.
	Code() Code

	// ClientMessage returns the text that is safe to return to the caller.
	ClientMessage() string
}

type moduleError struct {
	code Code
	err  error
}

// Code returns the error code.
func (e *moduleError) Code() Code {
	return e.code
}

// Error returns the code followed by the wrapped error, so printing this value
// prints the whole chain beneath it.
func (e *moduleError) Error() string {
	return string(e.code) + ": " + e.err.Error()
}

// ClientMessage returns the wrapped error's message, except for an internal
// failure, which always answers with a generic sentence. An error that wraps
// an internal failure anywhere in its chain counts as one, so a cause another
// service marked internal cannot resurface through a caller that classified
// the failure differently.
func (e *moduleError) ClientMessage() string {
	if e.code == CodeInternalFailureError || wrapsInternalFailure(e.err) {
		return internalFailureMessage
	}
	return e.err.Error()
}

// Unwrap unwraps an error.
func (e *moduleError) Unwrap() error {
	return e.err
}

// maxUnwrapDepth bounds how far the chain walk descends. A chain longer than
// this is either a cycle or pathological, and neither should be able to
// exhaust the stack.
const maxUnwrapDepth = 100

// wrapsInternalFailure reports whether any error in err's chain is a module
// error marked as an internal failure.
//
// It walks the chain by hand rather than using errors.As, which stops at the
// first module error it finds and would miss an internal failure wrapped
// further down. Both shapes of Unwrap are followed, so an internal failure
// combined through errors.Join is still found.
func wrapsInternalFailure(err error) bool {
	return wrapsInternalFailureWithin(err, maxUnwrapDepth)
}

// wrapsInternalFailureWithin is wrapsInternalFailure with a descent budget.
//
// Exhausting the budget answers true. Past that point the chain has not been
// fully inspected, and the safe answer is the one that keeps the cause hidden:
// a chain that deep yields the generic message rather than risking a leak.
func wrapsInternalFailureWithin(err error, budget int) bool {
	if budget == 0 {
		return true
	}
	if me, ok := err.(Error); ok && me.Code() == CodeInternalFailureError {
		return true
	}
	switch unwrapper := err.(type) {
	case interface{ Unwrap() error }:
		return wrapsInternalFailureWithin(unwrapper.Unwrap(), budget-1)
	case interface{ Unwrap() []error }:
		for _, wrapped := range unwrapper.Unwrap() {
			if wrapsInternalFailureWithin(wrapped, budget-1) {
				return true
			}
		}
	}
	return false
}

// NewInvalidInputError returns an error with CodeInvalidInputError.
//
// err's message is handed to the caller as-is, so it must already read as a
// complete sentence that is safe to expose.
func NewInvalidInputError(err error) error {
	return &moduleError{code: CodeInvalidInputError, err: err}
}

// NewNotFoundError returns an error with CodeNotFoundError.
//
// err's message is handed to the caller as-is, so it must already read as a
// complete sentence that is safe to expose.
func NewNotFoundError(err error) error {
	return &moduleError{code: CodeNotFoundError, err: err}
}

// NewInternalFailureError returns an error with CodeInternalFailureError.
//
// err is kept for Error and never reaches ClientMessage, so it is free to
// carry internal detail.
func NewInternalFailureError(err error) error {
	return &moduleError{code: CodeInternalFailureError, err: err}
}

// NewConflictError returns an error with CodeConflictError.
//
// err's message is handed to the caller as-is, so it must already read as a
// complete sentence that is safe to expose.
func NewConflictError(err error) error {
	return &moduleError{code: CodeConflictError, err: err}
}
