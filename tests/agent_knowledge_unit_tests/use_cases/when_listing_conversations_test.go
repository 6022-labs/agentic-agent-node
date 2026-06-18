package use_cases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	knowledge_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingConversationsTestingSuite struct {
	sut *use_cases.ListConversations

	mockRepository *knowledge_mocks.MockConversationMessageRepositoryInterface
}

func WhenListingConversationsBeforeEach(t *testing.T) *WhenListingConversationsTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := knowledge_mocks.NewMockConversationMessageRepositoryInterface(mockController)

	return &WhenListingConversationsTestingSuite{
		sut:            use_cases.NewListConversations(mockRepository),
		mockRepository: mockRepository,
	}
}

func TestWhenListingConversations(t *testing.T) {
	t.Parallel()

	t.Run("Given two conversations with messages", func(t *testing.T) {
		t.Parallel()

		first := time.Now().UTC().Truncate(time.Microsecond)
		second := first.Add(time.Minute)

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockRepository.EXPECT().ListConversationIds(gomock.Any()).Return([]string{"a", "b"}, nil)
			suite.mockRepository.EXPECT().ListByConversation(gomock.Any(), "a").Return([]models.ConversationMessage{
				model_builders.NewConversationMessageBuilder().WithConversationId("a").WithTimestamp(first).Build(),
				model_builders.NewConversationMessageBuilder().WithConversationId("a").WithTimestamp(second).Build(),
			}, nil)
			suite.mockRepository.EXPECT().ListByConversation(gomock.Any(), "b").Return([]models.ConversationMessage{
				model_builders.NewConversationMessageBuilder().WithConversationId("b").WithTimestamp(first).Build(),
			}, nil)
		}

		t.Run("Should summarise each conversation, preserving id order and first/last timestamps", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Len(t, resp.Conversations, 2)
			assert.Equal(t, "a", resp.Conversations[0].ConversationId)
			assert.Equal(t, int64(2), resp.Conversations[0].MessageCount)
			assert.Equal(t, first, resp.Conversations[0].FirstMessageAt)
			assert.Equal(t, second, resp.Conversations[0].LastMessageAt)
			assert.Equal(t, "b", resp.Conversations[1].ConversationId)
			assert.Equal(t, int64(1), resp.Conversations[1].MessageCount)
		})
	})

	t.Run("Given a conversation id with no messages", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockRepository.EXPECT().ListConversationIds(gomock.Any()).Return([]string{"empty"}, nil)
			suite.mockRepository.EXPECT().ListByConversation(gomock.Any(), "empty").Return(nil, nil)
		}

		t.Run("Should skip it", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Empty(t, resp.Conversations)
		})
	})

	t.Run("Given listing the ids fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockRepository.EXPECT().ListConversationIds(gomock.Any()).Return(nil, errors.New("db blew"))
		}

		t.Run("Should return InternalError carrying the repository message", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Contains(t, internalErr.Message, "db blew")
		})
	})
}
