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

		statusErr := &helpers.RuntimeStatusError{StatusCode: http.StatusInternalServerError, Code: "inference_failed"}

		t.Run("Should collapse it to unavailable", func(t *testing.T) {
			t.Parallel()

			err := helpers.MapRuntimeError(&helpers.RuntimeStatusError{StatusCode: http.StatusInternalServerError})

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
		})

		// The message of a 500 is the runtime's own err.Error() — driver and schema detail.
		t.Run("Should answer with the fixed message", func(t *testing.T) {
			t.Parallel()

			err := helpers.MapRuntimeError(&helpers.RuntimeStatusError{
				StatusCode: http.StatusInternalServerError,
				Code:       "inference_failed",
				Message:    "dial tcp 10.0.0.5:5432: connect: refused",
				Body:       `{"code":"inference_failed","message":"dial tcp 10.0.0.5:5432: connect: refused"}`,
			})

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
			assert.Equal(t, "the runtime is unavailable", unavailableErr.Message)
		})

		t.Run("Should keep the status error as the cause", func(t *testing.T) {
			t.Parallel()

			err := helpers.MapRuntimeError(statusErr)

			assert.ErrorIs(t, err, statusErr)
		})
	})

	t.Run("Given an error that is not a runtime status error", func(t *testing.T) {
		t.Parallel()

		transportErr := errors.New("dial tcp: connect: refused")

		// A transport error names the loopback address the runtime binds to.
		t.Run("Should not disclose the runtime's address", func(t *testing.T) {
			t.Parallel()

			err := helpers.MapRuntimeError(
				errors.New("Post \"http://127.0.0.1:5400/internal/responses\": dial tcp: connect: refused"),
			)

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
			assert.NotContains(t, unavailableErr.Message, "127.0.0.1")
			assert.NotContains(t, unavailableErr.Message, "5400")
		})

		t.Run("Should keep the error as the cause", func(t *testing.T) {
			t.Parallel()

			err := helpers.MapRuntimeError(transportErr)

			assert.ErrorIs(t, err, transportErr)
		})
	})
}
