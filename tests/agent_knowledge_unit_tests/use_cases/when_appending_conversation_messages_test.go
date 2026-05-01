package use_cases_test

import (
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	knowledge_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenAppendingConversationMessagesTestingSuite struct {
	sut *use_cases.AppendConversationMessages

	mockRepository *knowledge_mocks.MockConversationMessageRepositoryInterface
}

func WhenAppendingConversationMessagesBeforeEach(t *testing.T) *WhenAppendingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := knowledge_mocks.NewMockConversationMessageRepositoryInterface(mockController)

	sut := use_cases.NewAppendConversationMessages(mockRepository)

	return &WhenAppendingConversationMessagesTestingSuite{
		sut: sut,

		mockRepository: mockRepository,
	}
}

func TestWhenAppendingConversationMessages(t *testing.T) {
	t.Parallel()

	now := time.Now()
	conversationId := "conv-1"

	t.Run("Given an empty conversationId", func(t *testing.T) {
		t.Parallel()

		t.Run("Should no-op without touching the repository", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)

			message := model_builders.NewConversationMessageBuilder().WithTimestamp(now).Build()
			res, err := suite.sut.Execute("   ", []models.ConversationMessage{message})

			assert.NoError(t, err)
			assert.Empty(t, res)
		})
	})

	t.Run("Given an empty messages slice", func(t *testing.T) {
		t.Parallel()

		t.Run("Should no-op", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)

			res, err := suite.sut.Execute(conversationId, nil)

			assert.NoError(t, err)
			assert.Empty(t, res)
		})
	})

	t.Run("Given a list of messages", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockRepository.EXPECT().AppendMessages(gomock.Any()).DoAndReturn(func(messages []models.ConversationMessage) ([]models.ConversationMessage, error) {
				assert.Len(t, messages, 2)
				for _, m := range messages {
					assert.Equal(t, conversationId, m.ConversationId)
				}
				return messages, nil
			})
		}

		t.Run("Should stamp conversationId on every row before delegating", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			input := []models.ConversationMessage{
				model_builders.NewConversationMessageBuilder().WithAuthor("alice").WithRole("user").WithContent("hi").WithTimestamp(now).Build(),
				model_builders.NewConversationMessageBuilder().WithAuthor("bob").WithRole("assistant").WithContent("hello").WithTimestamp(now).Build(),
			}
			res, err := suite.sut.Execute(conversationId, input)

			assert.NoError(t, err)
			assert.Len(t, res, 2)
		})
	})

	t.Run("Given the repository fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockRepository.EXPECT().AppendMessages(gomock.Any()).Return(nil, errors.New("db down"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			message := model_builders.NewConversationMessageBuilder().WithTimestamp(now).Build()
			_, err := suite.sut.Execute(conversationId, []models.ConversationMessage{message})

			assert.ErrorContains(t, err, "db down")
		})
	})
}
