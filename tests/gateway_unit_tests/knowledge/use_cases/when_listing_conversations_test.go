package use_cases_test

import (
	"context"
	"errors"
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

type WhenListingConversationsTestingSuite struct {
	sut *use_cases.ListConversations

	mockClient *interfaces_mocks.MockConversationClientInterface
}

func WhenListingConversationsBeforeEach(t *testing.T) *WhenListingConversationsTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockConversationClientInterface(mockController)

	return &WhenListingConversationsTestingSuite{
		sut:        use_cases.NewListConversations(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenListingConversations(t *testing.T) {
	t.Parallel()

	t.Run("Given the client returns conversations", func(t *testing.T) {
		t.Parallel()

		now := time.Now().UTC().Truncate(time.Microsecond)

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockClient.EXPECT().
				ListConversations(gomock.Any(), uint(1), uint(10)).
				Return(&models.ExternalListConversationsPayload{
					Page:            1,
					ElementsPerPage: 10,
					TotalElements:   1,
					Conversations: []models.ExternalConversationPayload{
						{
							ConversationId: "a",
							Name:           lo.ToPtr("ERC20 mint audit"),
							MessageCount:   3,
							CreatedAt:      now,
							UpdatedAt:      now.Add(time.Minute),
						},
					},
				}, nil)
		}

		t.Run("Should map them onto the wire response", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), &requests.ListConversationsRequest{})

			assert.NoError(t, err)
			assert.Len(t, resp.Conversations, 1)
			assert.Equal(t, uint(1), resp.TotalElements)
			assert.Equal(t, "a", resp.Conversations[0].ConversationId)
			assert.Equal(t, "ERC20 mint audit", *resp.Conversations[0].Name)
			assert.Equal(t, int64(3), resp.Conversations[0].MessageCount)
			assert.Equal(t, now, resp.Conversations[0].CreatedAt)
			assert.Equal(t, now.Add(time.Minute), resp.Conversations[0].UpdatedAt)
		})
	})

	t.Run("Given the client fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockClient.EXPECT().
				ListConversations(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should return Unavailable carrying the message", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.ListConversationsRequest{})

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Contains(t, unavailableErr.Message, "runtime down")
		})
	})
}
