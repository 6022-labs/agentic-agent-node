package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingSignAgentCardRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a card object without a signatures member", func(t *testing.T) {
		t.Parallel()
		req := &requests.SignAgentCardRequest{Card: []byte(`{"name":"hermes","protocol":"6022"}`)}

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given a card that already carries a signatures member", func(t *testing.T) {
		t.Parallel()
		req := &requests.SignAgentCardRequest{Card: []byte(`{"name":"hermes","signatures":[{"protected":"x","signature":"y"}]}`)}

		t.Run("Should reject it, since the spec excludes that member from the signed content", func(t *testing.T) {
			t.Parallel()

			err := req.ValidateAndSanitize()

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "card", validationErr.Field)
		})
	})

	t.Run("Given a card that is not a JSON object", func(t *testing.T) {
		t.Parallel()
		req := &requests.SignAgentCardRequest{Card: []byte(`["not","an","object"]`)}

		t.Run("Should reject with a validation error on the card field", func(t *testing.T) {
			t.Parallel()

			err := req.ValidateAndSanitize()

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "card", validationErr.Field)
		})
	})

	t.Run("Given the JSON literal null", func(t *testing.T) {
		t.Parallel()
		req := &requests.SignAgentCardRequest{Card: []byte(`null`)}

		t.Run("Should reject it — unmarshalling null into a map does not error", func(t *testing.T) {
			t.Parallel()

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr)
		})
	})

	t.Run("Given an absent card", func(t *testing.T) {
		t.Parallel()
		req := &requests.SignAgentCardRequest{}

		t.Run("Should reject it rather than sign nothing", func(t *testing.T) {
			t.Parallel()

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr)
		})
	})
}
