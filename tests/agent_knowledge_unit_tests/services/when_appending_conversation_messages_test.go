package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/utils"
	knowledge_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenAppendingConversationMessagesTestingSuite struct {
	sut *services.AppendConversationMessages

	mockRepository *knowledge_mocks.MockConversationMessageRepositoryInterface
}

func WhenAppendingConversationMessagesBeforeEach(t *testing.T) *WhenAppendingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := knowledge_mocks.NewMockConversationMessageRepositoryInterface(mockController)

	sut := services.NewAppendConversationMessages(mockRepository)

	return &WhenAppendingConversationMessagesTestingSuite{
		sut:            sut,
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
			res, err := suite.sut.Append("   ", []models.ConversationMessage{message})

			assert.NoError(t, err)
			assert.Empty(t, res)
		})
	})

	t.Run("Given an empty messages slice", func(t *testing.T) {
		t.Parallel()

		t.Run("Should no-op", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)

			res, err := suite.sut.Append(conversationId, nil)

			assert.NoError(t, err)
			assert.Empty(t, res)
		})
	})

	t.Run("Given inputs whose hashes are all new", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockRepository.EXPECT().
				ListExistingHashes(gomock.Any()).
				Return([]string{}, nil)

			suite.mockRepository.EXPECT().
				AppendMessages(gomock.Any()).
				DoAndReturn(func(messages []models.ConversationMessage) ([]models.ConversationMessage, error) {
					assert.Len(t, messages, 2)
					for _, m := range messages {
						assert.Equal(t, conversationId, m.ConversationId)
						assert.NotEmpty(t, m.Hash)
					}
					return messages, nil
				})
		}

		t.Run("Should stamp conversationId, demojize, hash, and insert all", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			input := []models.ConversationMessage{
				model_builders.NewConversationMessageBuilder().WithRole("user").WithContent("hi 🔥").WithTimestamp(now).Build(),
				model_builders.NewConversationMessageBuilder().WithRole("assistant").WithContent("hello").WithTimestamp(now).Build(),
			}
			res, err := suite.sut.Append(conversationId, input)

			assert.NoError(t, err)
			assert.Len(t, res, 2)
			// First input had unicode emoji — content is demojized in the persisted row.
			assert.Equal(t, "hi :fire:", res[0].Content)
		})
	})

	t.Run("Given some inputs match existing hashes", func(t *testing.T) {
		t.Parallel()

		duplicateInput := model_builders.NewConversationMessageBuilder().WithRole("user").WithContent("dup").WithTimestamp(now).Build()
		duplicateInput.ConversationId = conversationId
		dupHash, _ := utils.ComputeConversationMessageHash(duplicateInput.Role, duplicateInput.Content, conversationId)

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockRepository.EXPECT().
				ListExistingHashes(gomock.Any()).
				Return([]string{dupHash}, nil)

			suite.mockRepository.EXPECT().
				AppendMessages(gomock.Any()).
				DoAndReturn(func(messages []models.ConversationMessage) ([]models.ConversationMessage, error) {
					if assert.Len(t, messages, 1) {
						assert.Equal(t, "fresh", messages[0].Content)
					}
					return messages, nil
				})
		}

		t.Run("Should drop dupes and only insert fresh rows", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			input := []models.ConversationMessage{
				model_builders.NewConversationMessageBuilder().WithRole("user").WithContent("dup").WithTimestamp(now).Build(),
				model_builders.NewConversationMessageBuilder().WithRole("user").WithContent("fresh").WithTimestamp(now).Build(),
			}
			res, err := suite.sut.Append(conversationId, input)

			assert.NoError(t, err)
			assert.Len(t, res, 1)
		})
	})

	t.Run("Given every input is already stored", func(t *testing.T) {
		t.Parallel()

		input := model_builders.NewConversationMessageBuilder().WithRole("user").WithContent("dup").WithTimestamp(now).Build()
		input.ConversationId = conversationId
		dupHash, _ := utils.ComputeConversationMessageHash(input.Role, input.Content, conversationId)

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockRepository.EXPECT().
				ListExistingHashes(gomock.Any()).
				Return([]string{dupHash}, nil)
			// no AppendMessages call expected
		}

		t.Run("Should short-circuit without inserting", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			res, err := suite.sut.Append(conversationId, []models.ConversationMessage{input})

			assert.NoError(t, err)
			assert.Empty(t, res)
		})
	})

	t.Run("Given ListExistingHashes fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockRepository.EXPECT().
				ListExistingHashes(gomock.Any()).
				Return(nil, errors.New("db blew"))
		}

		t.Run("Should propagate the error before inserting", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			input := model_builders.NewConversationMessageBuilder().WithTimestamp(now).Build()
			_, err := suite.sut.Append(conversationId, []models.ConversationMessage{input})

			assert.ErrorContains(t, err, "db blew")
		})
	})

	t.Run("Given AppendMessages fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockRepository.EXPECT().
				ListExistingHashes(gomock.Any()).
				Return([]string{}, nil)
			suite.mockRepository.EXPECT().
				AppendMessages(gomock.Any()).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			message := model_builders.NewConversationMessageBuilder().WithTimestamp(now).Build()
			_, err := suite.sut.Append(conversationId, []models.ConversationMessage{message})

			assert.ErrorContains(t, err, "db down")
		})
	})
}
