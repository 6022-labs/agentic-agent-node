package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingUpdateSelfToolRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given enabled is set", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			req := &requests.UpdateSelfToolRequest{Id: uuid.New(), Enabled: lo.ToPtr(false)}

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given enabled is absent", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject with a validation error on enabled", func(t *testing.T) {
			t.Parallel()

			req := &requests.UpdateSelfToolRequest{Id: uuid.New()}

			assert.ErrorContains(t, req.ValidateAndSanitize(), "enabled")
		})
	})
}
