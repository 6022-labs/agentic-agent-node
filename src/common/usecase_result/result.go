// Package usecase_result is the domain-neutral envelope every use case returns.
// It describes *what happened* in semantic terms (NotFound, Conflict, ...) so
// the domain layer never has to know about HTTP, fiber, or any transport.
//
// Transport adapters (MVC, RPC, ...) translate Outcome -> their own status.
package usecase_result

type Outcome int

const (
	OutcomeOk            Outcome = iota // happy path with body
	OutcomeCreated                      // resource created with body
	OutcomeNoContent                    // happy path, no body
	OutcomeInvalidInput                 // caller-supplied data is malformed/invalid
	OutcomeUnauthorized                 // caller failed to prove identity
	OutcomeForbidden                    // caller is known but not permitted
	OutcomeNotFound                     // requested resource does not exist
	OutcomeConflict                     // current state forbids the operation
	OutcomeUnavailable                  // upstream/dependency unavailable
	OutcomeInternalError                // unexpected internal failure
)

// Result is the envelope every use case returns.
//
// Body carries the success payload (or nil for content-less results).
// Code + Message describe failure outcomes; Body stays nil in that case.
type Result struct {
	Outcome Outcome
	Body    any
	Code    string
	Message string
}

func Ok(body any) *Result {
	return &Result{Outcome: OutcomeOk, Body: body}
}

func Created(body any) *Result {
	return &Result{Outcome: OutcomeCreated, Body: body}
}

func NoContent() *Result {
	return &Result{Outcome: OutcomeNoContent}
}

func InvalidInput(code, message string) *Result {
	return &Result{Outcome: OutcomeInvalidInput, Code: code, Message: message}
}

func Unauthorized(code, message string) *Result {
	return &Result{Outcome: OutcomeUnauthorized, Code: code, Message: message}
}

func Forbidden(code, message string) *Result {
	return &Result{Outcome: OutcomeForbidden, Code: code, Message: message}
}

func NotFound(code, message string) *Result {
	return &Result{Outcome: OutcomeNotFound, Code: code, Message: message}
}

func Conflict(code, message string) *Result {
	return &Result{Outcome: OutcomeConflict, Code: code, Message: message}
}

func Unavailable(code, message string) *Result {
	return &Result{Outcome: OutcomeUnavailable, Code: code, Message: message}
}

func InternalError(code, message string) *Result {
	return &Result{Outcome: OutcomeInternalError, Code: code, Message: message}
}
