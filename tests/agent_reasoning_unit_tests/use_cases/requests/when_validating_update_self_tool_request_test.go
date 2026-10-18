package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingUpdateSelfToolRequest(t *testing.T) {
	t.Parallel()

	generateUpdateSelfToolRequest := func() *requests.UpdateSelfToolRequest {
		enabled := true
		return &requests.UpdateSelfToolRequest{
			Id:      uuid.New(),
			Enabled: &enabled,
		}
	}

	t.Run("Given enabled is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the enabled field", func(t *testing.T) {
			t.Parallel()

			req := generateUpdateSelfToolRequest()
			req.Enabled = nil

			err := req.ValidateAndSanitize()

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "enabled", validationErr.Field)
		})
	})

	t.Run("Given enabled is set", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			req := generateUpdateSelfToolRequest()

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})
}
