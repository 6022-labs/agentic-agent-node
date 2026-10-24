package models_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenCheckingAgentOwnership(t *testing.T) {
	t.Parallel()

	callerId := strings.ToLower(common_tests.RandomEthAddress().Hex())

	t.Run("Given no owner", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not own the agent", func(t *testing.T) {
			t.Parallel()

			agent := model_builders.NewAgentBuilder().Build()
			agent.Owner = nil

			assert.False(t, agent.IsOwnedBy(callerId))
		})
	})

	t.Run("Given another owner", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not own the agent", func(t *testing.T) {
			t.Parallel()

			agent := model_builders.NewAgentBuilder().Build()

			assert.False(t, agent.IsOwnedBy(callerId))
		})
	})

	t.Run("Given the caller is the owner in a different case", func(t *testing.T) {
		t.Parallel()

		t.Run("Should own the agent", func(t *testing.T) {
			t.Parallel()

			agent := model_builders.NewAgentBuilder().WithOwner(strings.ToUpper(callerId)).Build()

			assert.True(t, agent.IsOwnedBy(callerId))
		})
	})
}
