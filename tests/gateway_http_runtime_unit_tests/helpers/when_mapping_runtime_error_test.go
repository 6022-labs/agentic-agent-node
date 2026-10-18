package helpers_test

import (
	"errors"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
	"github.com/stretchr/testify/assert"
)

func TestWhenMappingRuntimeError(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		t.Run("Should map it to a validation error with the runtime field", func(t *testing.T) {
			t.Parallel()

			err := helpers.MapRuntimeError(&helpers.RuntimeStatusError{
				StatusCode: http.StatusBadRequest,
				Field:      "conversationId",
				Message:    "cannot be empty",
			})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "conversationId", validationErr.Field)
			assert.Equal(t, "cannot be empty", validationErr.Message)
		})
	})

	t.Run("Given the runtime answers 401", func(t *testing.T) {
		t.Parallel()

		t.Run("Should map it to an unauthorized error with the runtime code", func(t *testing.T) {
			t.Parallel()

			err := helpers.MapRuntimeError(&helpers.RuntimeStatusError{
				StatusCode: http.StatusUnauthorized,
				Code:       "missing_caller_id",
				Message:    "an authenticated caller id is required",
			})

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "missing_caller_id", unauthorizedErr.Code)
		})
	})

	t.Run("Given the runtime answers 404", func(t *testing.T) {
		t.Parallel()

		t.Run("Should map it to a not-found error with the runtime code", func(t *testing.T) {
			t.Parallel()

			err := helpers.MapRuntimeError(&helpers.RuntimeStatusError{
				StatusCode: http.StatusNotFound,
				Code:       "conversation_not_found",
				Message:    "conversation not found",
			})

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "conversation_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given the runtime answers 409", func(t *testing.T) {
		t.Parallel()

		t.Run("Should map it to a conflict error with the runtime code", func(t *testing.T) {
			t.Parallel()

			err := helpers.MapRuntimeError(&helpers.RuntimeStatusError{
				StatusCode: http.StatusConflict,
				Code:       "agent_not_minted",
				Message:    "agent identity is not minted yet",
			})

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "agent_not_minted", conflictErr.Code)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		t.Run("Should collapse it to unavailable", func(t *testing.T) {
			t.Parallel()

			err := helpers.MapRuntimeError(&helpers.RuntimeStatusError{StatusCode: http.StatusInternalServerError})

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
		})
	})

	t.Run("Given an error that is not a runtime status error", func(t *testing.T) {
		t.Parallel()

		t.Run("Should map it to unavailable carrying its message", func(t *testing.T) {
			t.Parallel()

			err := helpers.MapRuntimeError(errors.New("dial tcp: connection refused"))

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
			assert.Contains(t, unavailableErr.Message, "connection refused")
		})
	})
}
