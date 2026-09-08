package moduleerror_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/rojbar/moduleerror/v2"
)

// internalFailureMessage is written out rather than referenced from the
// package, so these tests pin the text an internal failure actually exposes.
const internalFailureMessage = "an unexpected error occurred"

// newByCode builds the error a case describes. It keeps the table rows as
// data, instead of every row carrying a reference to its constructor.
func newByCode(t *testing.T, code moduleerror.Code, cause error) error {
	t.Helper()
	switch code {
	case moduleerror.CodeInvalidInputError:
		return moduleerror.NewInvalidInputError(cause)
	case moduleerror.CodeNotFoundError:
		return moduleerror.NewNotFoundError(cause)
	case moduleerror.CodeInternalFailureError:
		return moduleerror.NewInternalFailureError(cause)
	case moduleerror.CodeConflictError:
		return moduleerror.NewConflictError(cause)
	}
	t.Fatalf("newByCode: unknown code %q", code)
	return nil
}

func asModuleError(t *testing.T, err error) moduleerror.Error {
	t.Helper()
	var me moduleerror.Error
	if !errors.As(err, &me) {
		t.Fatalf("errors.As() found no moduleerror.Error in %v", err)
	}
	return me
}

func TestNew_textsPerCode(t *testing.T) {
	tests := []struct {
		name              string
		code              moduleerror.Code
		cause             string
		wantError         string
		wantClientMessage string
	}{
		{
			name:              "invalid_input_exposes_the_cause",
			code:              moduleerror.CodeInvalidInputError,
			cause:             "title must not be empty",
			wantError:         "INVALID_INPUT: title must not be empty",
			wantClientMessage: "title must not be empty",
		},
		{
			name:              "not_found_exposes_the_cause",
			code:              moduleerror.CodeNotFoundError,
			cause:             "epic not found",
			wantError:         "NOT_FOUND: epic not found",
			wantClientMessage: "epic not found",
		},
		{
			name:              "conflict_exposes_the_cause",
			code:              moduleerror.CodeConflictError,
			cause:             "goal already exists",
			wantError:         "CONFLICT: goal already exists",
			wantClientMessage: "goal already exists",
		},
		{
			name:              "internal_failure_hides_the_cause",
			code:              moduleerror.CodeInternalFailureError,
			cause:             "sqlite: database is locked",
			wantError:         "INTERNAL_FAILURE: sqlite: database is locked",
			wantClientMessage: internalFailureMessage,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := newByCode(t, tc.code, errors.New(tc.cause))

			me := asModuleError(t, err)

			if me.Code() != tc.code {
				t.Errorf("Code() = %q, want %q", me.Code(), tc.code)
			}
			if me.Error() != tc.wantError {
				t.Errorf("Error() = %q, want %q", me.Error(), tc.wantError)
			}
			if me.ClientMessage() != tc.wantClientMessage {
				t.Errorf("ClientMessage() = %q, want %q", me.ClientMessage(), tc.wantClientMessage)
			}
		})
	}
}

// nilUnwrapper is an error that advertises Unwrap but has nothing beneath it,
// the shape a third-party error can take.
type nilUnwrapper struct{}

func (nilUnwrapper) Error() string { return "wrapper with no cause" }
func (nilUnwrapper) Unwrap() error { return nil }

func TestClientMessage_internalFailureIsContagious(t *testing.T) {
	tests := []struct {
		name              string
		outerCode         moduleerror.Code
		innerCode         moduleerror.Code // empty wraps the cause directly
		join              bool
		innerUnwrapsToNil bool
		wantClientMessage string
	}{
		{
			name:              "plain_cause_is_exposed",
			outerCode:         moduleerror.CodeNotFoundError,
			wantClientMessage: "sqlite: database is locked",
		},
		{
			name:              "wrapping_a_non_internal_failure_exposes_it",
			outerCode:         moduleerror.CodeNotFoundError,
			innerCode:         moduleerror.CodeConflictError,
			wantClientMessage: "CONFLICT: sqlite: database is locked",
		},
		{
			name:              "wrapping_an_internal_failure_returns_the_generic_message",
			outerCode:         moduleerror.CodeNotFoundError,
			innerCode:         moduleerror.CodeInternalFailureError,
			wantClientMessage: internalFailureMessage,
		},
		{
			name:              "internal_failure_reached_through_errors_join_returns_the_generic_message",
			outerCode:         moduleerror.CodeNotFoundError,
			innerCode:         moduleerror.CodeInternalFailureError,
			join:              true,
			wantClientMessage: internalFailureMessage,
		},
		{
			name:              "errors_join_without_an_internal_failure_stays_exposed",
			outerCode:         moduleerror.CodeNotFoundError,
			innerCode:         moduleerror.CodeConflictError,
			join:              true,
			wantClientMessage: "unrelated failure\nCONFLICT: sqlite: database is locked",
		},
		{
			name:              "cause_that_unwraps_to_nothing_is_exposed",
			outerCode:         moduleerror.CodeNotFoundError,
			innerUnwrapsToNil: true,
			wantClientMessage: "wrapper with no cause",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			inner := error(errors.New("sqlite: database is locked"))
			if tc.innerUnwrapsToNil {
				inner = nilUnwrapper{}
			}
			if tc.innerCode != "" {
				inner = newByCode(t, tc.innerCode, inner)
			}
			if tc.join {
				inner = errors.Join(errors.New("unrelated failure"), inner)
			}

			got := asModuleError(t, newByCode(t, tc.outerCode, inner)).ClientMessage()

			if got != tc.wantClientMessage {
				t.Errorf("ClientMessage() = %q, want %q", got, tc.wantClientMessage)
			}
		})
	}
}

func TestNew_preservesTheChain(t *testing.T) {
	tests := []struct {
		name      string
		code      moduleerror.Code
		extraWrap bool
	}{
		{
			name: "errors_is_reaches_the_cause",
			code: moduleerror.CodeInternalFailureError,
		},
		{
			name:      "errors_as_finds_the_module_error_under_fmt_wrapping",
			code:      moduleerror.CodeNotFoundError,
			extraWrap: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cause := errors.New("sqlite: database is locked")
			err := newByCode(t, tc.code, cause)
			if tc.extraWrap {
				err = fmt.Errorf("creating goal: %w", err)
			}

			if !errors.Is(err, cause) {
				t.Errorf("errors.Is() did not reach the cause through %v", err)
			}
			if got := asModuleError(t, err).Code(); got != tc.code {
				t.Errorf("Code() = %q, want %q", got, tc.code)
			}
		})
	}
}

// cyclicError unwraps to itself, the shape an unbounded walk would follow
// forever.
type cyclicError struct{ next error }

func (c *cyclicError) Error() string { return "cyclic" }
func (c *cyclicError) Unwrap() error { return c.next }

func TestClientMessage_boundsTheChainWalk(t *testing.T) {
	tests := []struct {
		name        string
		depth       int
		cyclic      bool
		wantGeneric bool
	}{
		{
			name:        "chain_within_the_bound_is_fully_inspected",
			depth:       50,
			wantGeneric: false,
		},
		{
			name:        "chain_deeper_than_the_bound_hides_the_cause",
			depth:       200,
			wantGeneric: true,
		},
		{
			name:        "cyclic_chain_terminates_and_hides_the_cause",
			cyclic:      true,
			wantGeneric: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cause := error(errors.New("bottom"))
			if tc.cyclic {
				cycle := &cyclicError{}
				cycle.next = cycle
				cause = cycle
			}
			for i := 0; i < tc.depth; i++ {
				cause = fmt.Errorf("layer %d: %w", i, cause)
			}

			want := cause.Error()
			if tc.wantGeneric {
				want = internalFailureMessage
			}

			got := asModuleError(t, moduleerror.NewNotFoundError(cause)).ClientMessage()

			if got != want {
				t.Errorf("ClientMessage() = %q, want %q", got, want)
			}
		})
	}
}
