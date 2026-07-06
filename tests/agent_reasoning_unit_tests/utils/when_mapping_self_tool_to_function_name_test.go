package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhenMappingSelfToolToFunctionName(t *testing.T) {
	t.Parallel()

	t.Run("Given a self tool", func(t *testing.T) {
		t.Parallel()

		t.Run("Should prefix its function name with the self namespace", func(t *testing.T) {
			t.Parallel()

			name := utils.SelfToolToFunctionName(&models.Tool{Name: "get_registration"})

			assert.Equal(t, "self_get_registration", name)
		})
	})
}
