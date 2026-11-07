package use_cases_test

import (
	"context"
	"encoding/json"
	"errors"
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

	mockAgentRegistrationClient  *identity_mocks.MockAgentRegistrationClientInterface
	mockAgentCardSignatureClient *identity_mocks.MockAgentCardSignatureClientInterface
	mockPaymentReceiveRuleClient *payment_mocks.MockPaymentReceiveRuleClientInterface
	signedBody                   []byte
}

func WhenBuildingWellKnownAgent6022BeforeEach(
	t *testing.T,
	runtimeUrl, publicPrefix string,
) *WhenBuildingWellKnownAgent6022TestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRegistrationClient := identity_mocks.NewMockAgentRegistrationClientInterface(mockController)
	mockAgentCardSignatureClient := identity_mocks.NewMockAgentCardSignatureClientInterface(mockController)
	mockPaymentReceiveRuleClient := payment_mocks.NewMockPaymentReceiveRuleClientInterface(mockController)

	sut := use_cases.NewGetWellKnownAgent6022(
		mockAgentRegistrationClient,
		mockAgentCardSignatureClient,
		mockPaymentReceiveRuleClient,
		&identity_settings.AgentIdentitySettings{RuntimeUrl: runtimeUrl},
		&core_settings.GatewaySettings{PublicPrefix: publicPrefix},
	)

	return &WhenBuildingWellKnownAgent6022TestingSuite{
		sut:                          sut,
		mockAgentRegistrationClient:  mockAgentRegistrationClient,
		mockAgentCardSignatureClient: mockAgentCardSignatureClient,
		mockPaymentReceiveRuleClient: mockPaymentReceiveRuleClient,
	}
}

// expectCardSignature captures the exact bytes handed to the runtime — the spec's signed content.
func (s *WhenBuildingWellKnownAgent6022TestingSuite) expectCardSignature(ctx context.Context) {
	s.mockAgentCardSignatureClient.EXPECT().
		Sign(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, card []byte) (*identity_models.ExternalAgentCardSignaturePayload, error) {
			s.signedBody = card
			return &identity_models.ExternalAgentCardSignaturePayload{
				Protected: "protected",
				Signature: "signature",
			}, nil
		})
}

func TestWhenBuildingWellKnownAgent6022(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the registration client fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingWellKnownAgent6022TestingSuite) {
			suite.mockAgentRegistrationClient.EXPECT().
				Get(ctx).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "down", nil))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingWellKnownAgent6022BeforeEach(t, "https://agent.example.com", "/api")
			initSuite(suite)

			document, err := suite.sut.Execute(ctx)

			assert.Nil(t, document)
			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
		})
	})

	t.Run("Given the payment rule client fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingWellKnownAgent6022TestingSuite) {
			suite.mockAgentRegistrationClient.EXPECT().
				Get(ctx).
				Return(&identity_models.ExternalAgentRegistrationPayload{Name: "hermes"}, nil)
			suite.mockPaymentReceiveRuleClient.EXPECT().
				List(ctx).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "down", nil))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingWellKnownAgent6022BeforeEach(t, "https://agent.example.com", "/api")
			initSuite(suite)

			document, err := suite.sut.Execute(ctx)

			assert.Nil(t, document)
			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
		})
	})

	t.Run("Given the runtime cannot sign the card", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingWellKnownAgent6022TestingSuite) {
			suite.mockAgentRegistrationClient.EXPECT().
				Get(ctx).
				Return(&identity_models.ExternalAgentRegistrationPayload{Name: "hermes"}, nil)
			suite.mockPaymentReceiveRuleClient.EXPECT().
				List(ctx).
				Return(nil, nil)
			suite.mockAgentCardSignatureClient.EXPECT().
				Sign(ctx, gomock.Any()).
				Return(nil, apperrors.NewInternalError("agent_card_signing_failed", errors.New("vault unreachable")))
		}

		t.Run("Should fail rather than serve an unsigned document", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingWellKnownAgent6022BeforeEach(t, "https://agent.example.com", "/api")
			initSuite(suite)

			document, err := suite.sut.Execute(ctx)

			assert.Nil(t, document)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})

	t.Run("Given the payment policy does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingWellKnownAgent6022TestingSuite) {
			suite.mockAgentRegistrationClient.EXPECT().
				Get(ctx).
				Return(&identity_models.ExternalAgentRegistrationPayload{Name: "hermes"}, nil)
			suite.mockPaymentReceiveRuleClient.EXPECT().
				List(ctx).
				Return(nil, apperrors.NewNotFoundError("payment_receive_policy_not_found", "no policy"))
			suite.expectCardSignature(ctx)
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
			suite.mockPaymentReceiveRuleClient.EXPECT().
				List(ctx).
				Return(nil, apperrors.NewConflictError("agent_not_minted", "agent identity is not minted yet"))
			suite.expectCardSignature(ctx)
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

	t.Run("Given no payment rules exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingWellKnownAgent6022TestingSuite) {
			suite.mockAgentRegistrationClient.EXPECT().
				Get(ctx).
				Return(&identity_models.ExternalAgentRegistrationPayload{Name: "hermes"}, nil)
			suite.mockPaymentReceiveRuleClient.EXPECT().
				List(ctx).
				Return([]payment_models.ExternalPaymentReceiveRule{}, nil)
			suite.expectCardSignature(ctx)
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

	t.Run("Given a custom public prefix", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingWellKnownAgent6022TestingSuite) {
			suite.mockAgentRegistrationClient.EXPECT().
				Get(ctx).
				Return(&identity_models.ExternalAgentRegistrationPayload{Name: "hermes"}, nil)
			suite.mockPaymentReceiveRuleClient.EXPECT().
				List(ctx).
				Return(nil, nil)
			suite.expectCardSignature(ctx)
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

	t.Run("Given identity resolves and payment rules exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingWellKnownAgent6022TestingSuite) {
			suite.mockAgentRegistrationClient.EXPECT().
				Get(ctx).
				Return(&identity_models.ExternalAgentRegistrationPayload{
					Name:      " hermes ",
					EnsDomain: "hermes.agents6022.eth",
				}, nil)
			suite.mockPaymentReceiveRuleClient.EXPECT().
				List(ctx).
				Return([]payment_models.ExternalPaymentReceiveRule{
					{Method: "x402"},
					{Method: "x402"},
					{Method: "l402"},
				}, nil)
			suite.expectCardSignature(ctx)
		}

		t.Run("Should describe the agent from its registration", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingWellKnownAgent6022BeforeEach(t, "https://agent.example.com", "/api")
			initSuite(suite)

			document, err := suite.sut.Execute(ctx)

			assert.NoError(t, err)
			assert.Equal(t, "6022", document.Protocol)
			assert.Equal(t, "0.1", document.ProtocolVersion)
			assert.Equal(t, "hermes", document.Name)
			assert.Equal(t, "hermes.agents6022.eth", document.EnsDomain)
		})

		t.Run("Should expose the responses url under the public prefix", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingWellKnownAgent6022BeforeEach(t, "https://agent.example.com", "/api")
			initSuite(suite)

			document, err := suite.sut.Execute(ctx)

			assert.NoError(t, err)
			assert.Equal(t, "https://agent.example.com/api/responses", document.Endpoints.Responses)
		})

		t.Run("Should list the distinct rule methods", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingWellKnownAgent6022BeforeEach(t, "https://agent.example.com", "/api")
			initSuite(suite)

			document, err := suite.sut.Execute(ctx)

			assert.NoError(t, err)
			assert.Equal(t, []string{"l402", "x402"}, document.PaymentMethods)
		})

		t.Run("Should attach the signature the runtime returned", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingWellKnownAgent6022BeforeEach(t, "https://agent.example.com", "/api")
			initSuite(suite)

			document, err := suite.sut.Execute(ctx)

			assert.NoError(t, err)
			if assert.Len(t, document.Signatures, 1) {
				assert.Equal(t, "protected", document.Signatures[0].Protected)
				assert.Equal(t, "signature", document.Signatures[0].Signature)
			}
		})

		t.Run("Should sign a body with no signatures member, as the spec requires", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingWellKnownAgent6022BeforeEach(t, "https://agent.example.com", "/api")
			initSuite(suite)

			_, err := suite.sut.Execute(ctx)
			assert.NoError(t, err)

			members := map[string]json.RawMessage{}
			assert.NoError(t, json.Unmarshal(suite.signedBody, &members))
			assert.NotContains(t, members, "signatures")
			assert.Contains(t, members, "protocol")
		})
	})
}
