package use_cases_test

import (
	"context"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	core_settings "github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases"
	identity_models "github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
	identity_settings "github.com/6022-labs/agentic-agent-node/src/gateway/identity/settings"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/identity_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenBuildingWellKnownAgentCardTestingSuite struct {
	sut *use_cases.GetWellKnownAgentCard

	mockAgentRegistrationClient *identity_mocks.MockAgentRegistrationClientInterface
}

func WhenBuildingWellKnownAgentCardBeforeEach(t *testing.T, runtimeUrl, publicPrefix string) *WhenBuildingWellKnownAgentCardTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRegistrationClient := identity_mocks.NewMockAgentRegistrationClientInterface(mockController)

	sut := use_cases.NewGetWellKnownAgentCard(
		mockAgentRegistrationClient,
		&identity_settings.AgentIdentitySettings{RuntimeUrl: runtimeUrl},
		&core_settings.GatewaySettings{PublicPrefix: publicPrefix},
	)

	return &WhenBuildingWellKnownAgentCardTestingSuite{
		sut:                         sut,
		mockAgentRegistrationClient: mockAgentRegistrationClient,
	}
}

func TestWhenBuildingWellKnownAgentCard(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given identity resolves", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingWellKnownAgentCardTestingSuite) {
			suite.mockAgentRegistrationClient.EXPECT().
				Get(ctx).
				Return(&identity_models.ExternalAgentRegistrationPayload{Name: "hermes"}, nil)
		}

		t.Run("Should build the card from the agent name with url = runtime_url + public prefix + /a2a", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingWellKnownAgentCardBeforeEach(t, "https://agent.example.com", "/api")
			initSuite(suite)

			card, err := suite.sut.Execute(ctx)

			assert.NoError(t, err)
			assert.Equal(t, "hermes", card.Name)
			assert.False(t, card.Capabilities.Streaming)
			if assert.Len(t, card.SupportedInterfaces, 1) {
				assert.Equal(t, "https://agent.example.com/api/a2a", card.SupportedInterfaces[0].Url)
				assert.Equal(t, "JSONRPC", card.SupportedInterfaces[0].ProtocolBinding)
				assert.Equal(t, "1.0", card.SupportedInterfaces[0].ProtocolVersion)
			}
			if assert.Len(t, card.Skills, 1) {
				assert.Equal(t, "chat", card.Skills[0].Id)
				assert.Equal(t, "hermes", card.Skills[0].Name)
			}
		})
	})

	t.Run("Given the registration client fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingWellKnownAgentCardTestingSuite) {
			suite.mockAgentRegistrationClient.EXPECT().
				Get(ctx).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "down"))
		}

		t.Run("Should propagate the error instead of serving a placeholder card", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingWellKnownAgentCardBeforeEach(t, "https://agent.example.com", "/api")
			initSuite(suite)

			card, err := suite.sut.Execute(ctx)

			assert.Nil(t, card)
			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
		})
	})

	t.Run("Given a custom public prefix", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingWellKnownAgentCardTestingSuite) {
			suite.mockAgentRegistrationClient.EXPECT().
				Get(ctx).
				Return(&identity_models.ExternalAgentRegistrationPayload{Name: "hermes"}, nil)
		}

		t.Run("Should honor it (the gateway never hardcodes the reverse-proxy mount)", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingWellKnownAgentCardBeforeEach(t, "https://agent.example.com", "/gw")
			initSuite(suite)

			card, err := suite.sut.Execute(ctx)

			assert.NoError(t, err)
			if assert.Len(t, card.SupportedInterfaces, 1) {
				assert.Equal(t, "https://agent.example.com/gw/a2a", card.SupportedInterfaces[0].Url)
			}
		})
	})
}
