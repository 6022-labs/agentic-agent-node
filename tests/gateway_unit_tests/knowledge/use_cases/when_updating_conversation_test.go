package use_cases_test

import (
	"context"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/knowledge_mocks/services_mocks/interfaces_mocks"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingConversationTestingSuite struct {
	sut *use_cases.UpdateConversation

	mockClient *interfaces_mocks.MockConversationClientInterface
}

func WhenUpdatingConversationBeforeEach(t *testing.T) *WhenUpdatingConversationTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockConversationClientInterface(mockController)

	return &WhenUpdatingConversationTestingSuite{
		sut:        use_cases.NewUpdateConversation(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenUpdatingConversation(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Microsecond)

	t.Run("Given the route id is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError without calling the runtime", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdateConversationRequest{
				ConversationId: "  ",
				Name:           "Solidity audits",
			})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Contains(t, err.Error(), "conversationId")
		})
	})

	t.Run("Given the runtime does not know the conversation", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				UpdateConversation(gomock.Any(), "ghost", gomock.Any()).
				Return(nil, apperrors.NewNotFoundError("conversation_not_found", "conversation not found"))
		}

		t.Run("Should surface the not-found unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdateConversationRequest{
				ConversationId: "ghost",
				Name:           "Solidity audits",
			})

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
		})
	})

	t.Run("Given the runtime renames the session", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				UpdateConversation(gomock.Any(), "c1", gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					_ string,
					body *models.ExternalUpdateConversationPayload,
				) (*models.ExternalConversationPayload, error) {
					assert.Equal(t, "Solidity audits", body.Name)
					return &models.ExternalConversationPayload{
						ConversationId: "c1",
						Name:           lo.ToPtr("Solidity audits"),
						CreatedAt:      now,
						UpdatedAt:      now,
					}, nil
				})
		}

		t.Run("Should trim the route id and map the payload onto the wire response", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), &requests.UpdateConversationRequest{
				ConversationId: " c1 ",
				Name:           "  Solidity   audits ",
			})

			assert.NoError(t, err)
			assert.Equal(t, "Solidity audits", *resp.Name)
			assert.Equal(t, now, resp.UpdatedAt)
		})
	})
}
