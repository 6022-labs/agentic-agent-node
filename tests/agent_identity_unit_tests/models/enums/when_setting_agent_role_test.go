package enums_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/stretchr/testify/assert"
)

func TestWhenSettingAgentRole(t *testing.T) {
	t.Parallel()

	t.Run("Given a known role in canonical casing", func(t *testing.T) {
		t.Parallel()
		var role enums.AgentRole

		t.Run("Should parse and assign it", func(t *testing.T) {
			err := role.Set("expert")
			assert.NoError(t, err)
			assert.Equal(t, enums.AgentRoleExpert, role)
		})
	})

	t.Run("Given a known role in mixed casing", func(t *testing.T) {
		t.Parallel()
		var role enums.AgentRole

		t.Run("Should normalize and assign it", func(t *testing.T) {
			err := role.Set("FaCiLiTaToR")
			assert.NoError(t, err)
			assert.Equal(t, enums.AgentRoleFacilitator, role)
		})
	})

	t.Run("Given an unknown role", func(t *testing.T) {
		t.Parallel()
		var role enums.AgentRole

		t.Run("Should reject with an error mentioning the value", func(t *testing.T) {
			err := role.Set("wizard")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "wizard")
		})
	})

	t.Run("Given the full list of roles", func(t *testing.T) {
		t.Parallel()
		roles := enums.GetAllAgentRoles()

		t.Run("Should contain every known role and exclude unknowns", func(t *testing.T) {
			assert.True(t, roles.Contains(enums.AgentRoleClone))
			assert.True(t, roles.Contains(enums.AgentRoleHuman))
			assert.True(t, roles.Contains(enums.AgentRoleExpert))
			assert.True(t, roles.Contains(enums.AgentRoleFacilitator))
			assert.False(t, roles.Contains(enums.AgentRole("wizard")))
		})
	})
}
