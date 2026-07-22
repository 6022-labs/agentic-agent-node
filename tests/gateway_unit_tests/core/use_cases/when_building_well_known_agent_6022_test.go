package use_cases_test

import (
	"context"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	core_settings "github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases"
	identity_models "github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
	identity_settings "github.com/6022-labs/agentic-agent-node/src/gateway/identity/settings"
	payment_models "github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/identity_mocks/services_mocks/interfaces_mocks"
	payment_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/payment_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenBuildingWellKnownAgent6022TestingSuite struct {
	sut *use_cases.GetWellKnownAgent6022

	mockAgentRegistrationClient *identity_mocks.MockAgentRegistrationClientInterface
	mockPaymentRuleClient       *payment_mocks.MockPaymentRuleClientInterface
}

func WhenBuildingWellKnownAgent6022BeforeEach(t *testing.T, runtimeUrl, publicPrefix string) *WhenBuildingWellKnownAgent6022TestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRegistrationClient := identity_mocks.NewMockAgentRegistrationClientInterface(mockController)
	mockPaymentRuleClient := payment_mocks.NewMockPaymentRuleClientInterface(mockController)

	sut := use_cases.NewGetWellKnownAgent6022(
		mockAgentRegistrationClient,
		mockPaymentRuleClient,
		&identity_settings.AgentIdentitySettings{RuntimeUrl: runtimeUrl},
		&core_settings.GatewaySettings{PublicPrefix: publicPrefix},
	)

	return &WhenBuildingWellKnownAgent6022TestingSuite{
		sut:                         sut,
		mockAgentRegistrationClient: mockAgentRegistrationClient,
		mockPaymentRuleClient:       mockPaymentRuleClient,
	}
}

func TestWhenBuildingWellKnownAgent6022(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given identity resolves and payment rules exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingWellKnownAgent6022TestingSuite) {
			suite.mockAgentRegistrationClient.EXPECT().
				Get(ctx).
				Return(&identity_models.ExternalAgentRegistrationPayload{
					Name:      " hermes ",
					EnsDomain: "hermes.agents6022.eth",
				}, nil)
			suite.mockPaymentRuleClient.EXPECT().
				List(ctx).
				Return([]payment_models.ExternalPaymentRule{
					{Method: "x402"},
					{Method: "x402"},
					{Method: "l402"},
				}, nil)
		}

		t.Run("Should build the document with responses url and the distinct rule methods", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingWellKnownAgent6022BeforeEach(t, "https://agent.example.com", "/api")
			initSuite(suite)

			document, err := suite.sut.Execute(ctx)

			assert.NoError(t, err)
			assert.Equal(t, "6022", document.Protocol)
			assert.Equal(t, "0.1", document.ProtocolVersion)
			assert.Equal(t, "hermes", document.Name)
			assert.Equal(t, "hermes.agents6022.eth", document.EnsDomain)
			assert.Equal(t, "https://agent.example.com/api/responses", document.Endpoints.Responses)
			assert.Equal(t, []string{"x402", "l402"}, document.PaymentMethods)
		})
	})

	t.Run("Given no payment rules are configured", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingWellKnownAgent6022TestingSuite) {
			suite.mockAgentRegistrationClient.EXPECT().
				Get(ctx).
				Return(&identity_models.ExternalAgentRegistrationPayload{Name: "hermes"}, nil)
			suite.mockPaymentRuleClient.EXPECT().
				List(ctx).
				Return([]payment_models.ExternalPaymentRule{}, nil)
		}

		t.Run("Should serve the document without payment methods", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingWellKnownAgent6022BeforeEach(t, "https://agent.example.com", "/api")
			initSuite(suite)

			document, err := suite.sut.Execute(ctx)

			assert.NoError(t, err)
			assert.Empty(t, document.PaymentMethods)
			assert.Empty(t, document.EnsDomain)
		})
	})

	t.Run("Given the payment policy does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingWellKnownAgent6022TestingSuite) {
			suite.mockAgentRegistrationClient.EXPECT().
				Get(ctx).
				Return(&identity_models.ExternalAgentRegistrationPayload{Name: "hermes"}, nil)
			suite.mockPaymentRuleClient.EXPECT().
				List(ctx).
				Return(nil, apperrors.NewNotFoundError("payment_policy_not_found", "no policy"))
		}

		t.Run("Should serve the document without payment methods rather than fail", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingWellKnownAgent6022BeforeEach(t, "https://agent.example.com", "/api")
			initSuite(suite)

			document, err := suite.sut.Execute(ctx)

			assert.NoError(t, err)
			assert.Empty(t, document.PaymentMethods)
		})
	})

	t.Run("Given the agent is not minted yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingWellKnownAgent6022TestingSuite) {
			suite.mockAgentRegistrationClient.EXPECT().
				Get(ctx).
				Return(&identity_models.ExternalAgentRegistrationPayload{Name: "hermes"}, nil)
			suite.mockPaymentRuleClient.EXPECT().
				List(ctx).
				Return(nil, apperrors.NewConflictError("agent_not_minted", "agent identity is not minted yet"))
		}

		t.Run("Should serve the document without payment methods rather than fail", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingWellKnownAgent6022BeforeEach(t, "https://agent.example.com", "/api")
			initSuite(suite)

			document, err := suite.sut.Execute(ctx)

			assert.NoError(t, err)
			assert.Empty(t, document.PaymentMethods)
		})
	})

	t.Run("Given the payment rule client fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingWellKnownAgent6022TestingSuite) {
			suite.mockAgentRegistrationClient.EXPECT().
				Get(ctx).
				Return(&identity_models.ExternalAgentRegistrationPayload{Name: "hermes"}, nil)
			suite.mockPaymentRuleClient.EXPECT().
				List(ctx).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "down"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingWellKnownAgent6022BeforeEach(t, "https://agent.example.com", "/api")
			initSuite(suite)

			document, err := suite.sut.Execute(ctx)

			assert.Nil(t, document)
			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
		})
	})

	t.Run("Given the registration client fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingWellKnownAgent6022TestingSuite) {
			suite.mockAgentRegistrationClient.EXPECT().
				Get(ctx).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "down"))
		}

		t.Run("Should propagate the error instead of serving a placeholder document", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingWellKnownAgent6022BeforeEach(t, "https://agent.example.com", "/api")
			initSuite(suite)

			document, err := suite.sut.Execute(ctx)

			assert.Nil(t, document)
			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
		})
	})

	t.Run("Given a custom public prefix", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingWellKnownAgent6022TestingSuite) {
			suite.mockAgentRegistrationClient.EXPECT().
				Get(ctx).
				Return(&identity_models.ExternalAgentRegistrationPayload{Name: "hermes"}, nil)
			suite.mockPaymentRuleClient.EXPECT().
				List(ctx).
				Return(nil, nil)
		}

		t.Run("Should honor it", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingWellKnownAgent6022BeforeEach(t, "https://agent.example.com", "/gw")
			initSuite(suite)

			document, err := suite.sut.Execute(ctx)

			assert.NoError(t, err)
			assert.Equal(t, "https://agent.example.com/gw/responses", document.Endpoints.Responses)
		})
	})
}
