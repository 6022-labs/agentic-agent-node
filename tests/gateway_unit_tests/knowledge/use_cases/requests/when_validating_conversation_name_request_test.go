package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingConversationNameRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a blank name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the name field", func(t *testing.T) {
			t.Parallel()

			req := &requests.ConversationNameRequest{Name: "   "}

			assert.ErrorContains(t, req.ValidateAndSanitize(), "name")
		})
	})

	t.Run("Given a padded name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should collapse it", func(t *testing.T) {
			t.Parallel()

			req := &requests.ConversationNameRequest{Name: "  ERC20   mint audit "}

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, "ERC20 mint audit", req.Name)
		})
	})
}
