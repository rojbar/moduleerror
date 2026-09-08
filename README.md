# moduleerror [![Go Reference](https://pkg.go.dev/badge/github.com/rojbar/moduleerror/v2.svg)](https://pkg.go.dev/github.com/rojbar/moduleerror/v2) ![build](https://github.com/rojbar/moduleerror/actions/workflows/build.yml/badge.svg) [![codecov](https://codecov.io/github/rojbar/moduleerror/graph/badge.svg?token=41C18OIZ1Q)](https://codecov.io/github/rojbar/moduleerror)

moduleerror implements common module error logic.

An error carries a `Code` that classifies the failure without naming any
transport, so the same value can become an HTTP status, an MCP tool failure or
a gRPC code at whichever boundary handles it.

```sh
go get github.com/rojbar/moduleerror/v2
```

## Two texts, two audiences

`Error` is for operators and always includes the whole wrapped chain.
`ClientMessage` is what is safe to hand back to the caller.

```go
err := moduleerror.NewInternalFailureError(
	fmt.Errorf("inserting goal row: %w", errors.New("sqlite: database is locked")))

err.Error()
// INTERNAL_FAILURE: inserting goal row: sqlite: database is locked

err.(moduleerror.Error).ClientMessage()
// an unexpected error occurred
```

For every other code the caller sees the cause, because it is already written
in domain vocabulary:

```go
err := moduleerror.NewInvalidInputError(errors.New("title must not be empty"))

err.Error()          // INVALID_INPUT: title must not be empty
ClientMessage()      // title must not be empty
```

That split is what lets one service call another. The caller logs the whole
chain the callee produced, while the client of the outermost service only ever
sees the generic message. An error that wraps an internal failure anywhere in
its chain counts as one, so a cause another service marked internal cannot
resurface through a caller that classified the failure differently.

## Writing the cause

For every code except `INTERNAL_FAILURE`, the cause's message reaches the
caller as-is, so it must already read on its own and be safe to expose. An
internal failure's cause never reaches the caller, so it is free to carry
driver output, identifiers or anything else useful in a log.

## At a boundary

Classify once, in a shared helper, not in every handler.

```go
var me moduleerror.Error
if errors.As(err, &me) {
	status, msg = statusByCode[me.Code()], me.ClientMessage()
} else {
	status, msg = http.StatusInternalServerError, "internal error"
}
```

## Codes

| Code | Meaning |
|---|---|
| `INVALID_INPUT` | the caller provided something invalid |
| `NOT_FOUND` | the resource does not exist |
| `CONFLICT` | the request conflicts with current state |
| `INTERNAL_FAILURE` | anything else; the cause is never exposed |
