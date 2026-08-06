package use_cases_test

import (
	"context"
	"errors"
	"testing"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSigningAgentCardTestingSuite struct {
	sut *use_cases.SignAgentCard

	mockAgentCardSigner *identity_mocks.MockAgentCardSignerInterface
}

func WhenSigningAgentCardBeforeEach(t *testing.T) *WhenSigningAgentCardTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentCardSigner := identity_mocks.NewMockAgentCardSignerInterface(mockController)

	return &WhenSigningAgentCardTestingSuite{
		sut:                 use_cases.NewSignAgentCard(mockAgentCardSigner),
		mockAgentCardSigner: mockAgentCardSigner,
	}
}

func TestWhenSigningAgentCard(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const card = `{"name":"hermes","protocol":"6022"}`

	t.Run("Given a card without a signatures member", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSigningAgentCardTestingSuite) {
			suite.mockAgentCardSigner.EXPECT().
				Sign(ctx, []byte(card)).
				Return(&identity_interfaces.AgentCardSignature{Protected: "protected", Signature: "signature"}, nil)
		}

		t.Run("Should return the detached JWS", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningAgentCardBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, &requests.SignAgentCardRequest{Card: []byte(card)})

			assert.NoError(t, err)
			if assert.NotNil(t, response) {
				assert.Equal(t, "protected", response.Protected)
				assert.Equal(t, "signature", response.Signature)
			}
		})
	})

	t.Run("Given the request fails validation", func(t *testing.T) {
		t.Parallel()

		t.Run("Should surface it without calling the signer", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningAgentCardBeforeEach(t)
			suite.mockAgentCardSigner.EXPECT().Sign(gomock.Any(), gomock.Any()).Times(0)

			signed := `{"name":"hermes","signatures":[{"protected":"x","signature":"y"}]}`
			response, err := suite.sut.Execute(ctx, &requests.SignAgentCardRequest{Card: []byte(signed)})

			assert.Nil(t, response)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the signer fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSigningAgentCardTestingSuite) {
			suite.mockAgentCardSigner.EXPECT().
				Sign(ctx, gomock.Any()).
				Return(nil, errors.New("vault unreachable"))
		}

		t.Run("Should surface an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningAgentCardBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, &requests.SignAgentCardRequest{Card: []byte(card)})

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})
}
