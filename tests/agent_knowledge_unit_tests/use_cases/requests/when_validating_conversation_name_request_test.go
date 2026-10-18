package requests_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingConversationNameRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a blank name", func(t *testing.T) {
		t.Parallel()
		req := &requests.ConversationNameRequest{Name: "   "}

		t.Run("Should reject on the name field", func(t *testing.T) {
			t.Parallel()

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "name", validationErr.Field)
		})
	})

	t.Run("Given a name longer than the column", func(t *testing.T) {
		t.Parallel()
		req := &requests.ConversationNameRequest{Name: strings.Repeat("a", requests.ConversationNameMaxLength+1)}

		t.Run("Should reject on the name field", func(t *testing.T) {
			t.Parallel()

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "name", validationErr.Field)
		})
	})

	t.Run("Given a uuid-shaped name", func(t *testing.T) {
		t.Parallel()
		req := &requests.ConversationNameRequest{Name: uuid.NewString()}

		t.Run("Should reject on the uuid rule", func(t *testing.T) {
			t.Parallel()

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "name", validationErr.Field)
			assert.ErrorContains(t, validationErr, "uuid")
		})
	})

	t.Run("Given a name of exactly the maximum length", func(t *testing.T) {
		t.Parallel()
		req := &requests.ConversationNameRequest{Name: strings.Repeat("a", requests.ConversationNameMaxLength)}

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given a padded name", func(t *testing.T) {
		t.Parallel()
		req := &requests.ConversationNameRequest{Name: "  ERC20   mint audit "}

		t.Run("Should collapse it", func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, "ERC20 mint audit", req.Name)
		})
	})
}
