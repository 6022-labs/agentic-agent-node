package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingChainIdRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a blank chain id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the chainId field", func(t *testing.T) {
			t.Parallel()

			req := &requests.ChainIdRequest{ChainId: "  "}

			assert.ErrorContains(t, req.ValidateAndSanitize(), "chainId")
		})
	})

	t.Run("Given a padded chain id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should trim it", func(t *testing.T) {
			t.Parallel()

			req := &requests.ChainIdRequest{ChainId: " 80002 "}

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, "80002", req.ChainId)
		})
	})
}
