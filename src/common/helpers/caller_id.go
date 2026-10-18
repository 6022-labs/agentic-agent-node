package helpers

import (
	"context"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// CallerIdMaxLength matches the conversations.caller_id column, in characters.
const CallerIdMaxLength = 128

// CallerIdHeader carries the authenticated caller id from the gateway to the runtime.
const CallerIdHeader = "X-Caller-Id"

var callerIdRule = "must be a non-empty caller id of at most " + strconv.Itoa(CallerIdMaxLength) + " characters"

type callerIdContextKey struct{}

// ContextWithCallerId attaches the request's caller id, normalized.
func ContextWithCallerId(ctx context.Context, callerId string) context.Context {
	return context.WithValue(ctx, callerIdContextKey{}, NormalizeCallerId(callerId))
}

// CallerIdFromContext reports false when no caller id was attached.
func CallerIdFromContext(ctx context.Context) (string, bool) {
	callerId, ok := ctx.Value(callerIdContextKey{}).(string)
	return callerId, ok && callerId != ""
}

// RequireCallerId fails Unauthorized without a caller id and Validation when it overflows the column.
func RequireCallerId(ctx context.Context) (string, error) {
	callerId, ok := CallerIdFromContext(ctx)
	if !ok {
		return "", errors.NewUnauthorizedError("missing_caller_id", "an authenticated caller id is required")
	}
	if !IsCallerId(callerId) {
		return "", errors.NewValidationError("callerId", callerIdRule)
	}
	return callerId, nil
}

// NormalizeCallerId lowercases: MySQL/MariaDB compare ids case-insensitively, so every driver must.
func NormalizeCallerId(callerId string) string {
	return strings.ToLower(strings.TrimSpace(callerId))
}

// IsCallerId accepts any non-blank opaque id that fits the column.
func IsCallerId(callerId string) bool {
	return strings.TrimSpace(callerId) != "" && utf8.RuneCountInString(callerId) <= CallerIdMaxLength
}
