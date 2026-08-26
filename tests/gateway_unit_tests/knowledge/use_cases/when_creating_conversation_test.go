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

type WhenCreatingConversationTestingSuite struct {
	sut *use_cases.CreateConversation

	mockClient *interfaces_mocks.MockConversationClientInterface
}

func WhenCreatingConversationBeforeEach(t *testing.T) *WhenCreatingConversationTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockConversationClientInterface(mockController)

	return &WhenCreatingConversationTestingSuite{
		sut:        use_cases.NewCreateConversation(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenCreatingConversation(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Microsecond)

	t.Run("Given the runtime creates the session", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				CreateConversation(gomock.Any(), gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					body *models.ExternalCreateConversationPayload,
				) (*models.ExternalConversationPayload, error) {
					assert.Equal(t, "ERC20 mint audit", body.Name)
					return &models.ExternalConversationPayload{
						ConversationId: "c1",
						Name:           lo.ToPtr("ERC20 mint audit"),
						CreatedAt:      now,
						UpdatedAt:      now,
					}, nil
				})
		}

		t.Run("Should sanitize the name and map the payload onto the wire response", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingConversationBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), &requests.CreateConversationRequest{
				Name: "  ERC20   mint audit ",
			})

			assert.NoError(t, err)
			assert.Equal(t, "c1", resp.ConversationId)
			assert.Equal(t, "ERC20 mint audit", *resp.Name)
			assert.Equal(t, now, resp.CreatedAt)
		})
	})

	t.Run("Given a request that fails validation", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError without calling the runtime", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingConversationBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), &requests.CreateConversationRequest{Name: "   "})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the runtime rejects the name", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				CreateConversation(gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NewConflictError("conversation_name_taken", "already used"))
		}

		t.Run("Should surface the conflict unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.CreateConversationRequest{Name: "taken"})

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "conversation_name_taken", conflictErr.Code)
		})
	})
}
