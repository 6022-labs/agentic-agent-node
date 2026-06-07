package services_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/settings"
	"github.com/stretchr/testify/assert"
)

type WhenProvidingRuntimeRecordsTestingSuite struct {
	sut *services.RuntimeRecordsProvider
}

func WhenProvidingRuntimeRecordsBeforeEach() *WhenProvidingRuntimeRecordsTestingSuite {
	sut := services.NewRuntimeRecordsProvider(&settings.AgentIdentitySettings{RuntimeUrl: runtimeUrl})

	return &WhenProvidingRuntimeRecordsTestingSuite{sut: sut}
}

func TestWhenProvidingRuntimeRecords(t *testing.T) {
	t.Parallel()

	t.Run("Given a configured runtime url", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenProvidingRuntimeRecordsTestingSuite) {}

		t.Run("Should return a single url text record carrying it", func(t *testing.T) {
			t.Parallel()

			suite := WhenProvidingRuntimeRecordsBeforeEach()
			initSuite(suite)

			records := suite.sut.RuntimeRecords()

			assert.Equal(t, []identity_interfaces.EnsRecord{{Key: "url", Value: runtimeUrl}}, records)
		})
	})
}
