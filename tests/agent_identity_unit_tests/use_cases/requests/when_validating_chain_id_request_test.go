package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingChainIdRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a blank chain id", func(t *testing.T) {
		t.Parallel()
		req := &requests.ChainIdRequest{ChainId: "  "}

		t.Run("Should reject on the chainId field", func(t *testing.T) {
			t.Parallel()

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "chainId", validationErr.Field)
		})
	})

	t.Run("Given a padded chain id", func(t *testing.T) {
		t.Parallel()
		req := &requests.ChainIdRequest{ChainId: " 80002 "}

		t.Run("Should trim it", func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, "80002", req.ChainId)
		})
	})
}
