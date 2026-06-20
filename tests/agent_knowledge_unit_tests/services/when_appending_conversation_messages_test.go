package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/utils"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	knowledge_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenAppendingConversationMessagesTestingSuite struct {
	sut *services.MessageAppender

	mockRepository             *knowledge_mocks.MockConversationMessageRepositoryInterface
	mockConversationRepository *knowledge_mocks.MockConversationRepositoryInterface
	mockAgentRepository        *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenAppendingConversationMessagesBeforeEach(t *testing.T) *WhenAppendingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := knowledge_mocks.NewMockConversationMessageRepositoryInterface(mockController)
	mockConversationRepository := knowledge_mocks.NewMockConversationRepositoryInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	sut := services.NewMessageAppender(mockRepository, mockConversationRepository, mockAgentRepository)

	return &WhenAppendingConversationMessagesTestingSuite{
		sut:                        sut,
		mockRepository:             mockRepository,
		mockConversationRepository: mockConversationRepository,
		mockAgentRepository:        mockAgentRepository,
	}
}

func (suite *WhenAppendingConversationMessagesTestingSuite) expectAgentResolutionAndEnsure() {
	suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
	suite.mockConversationRepository.EXPECT().Ensure(gomock.Any(), gomock.Any()).Return(nil)
}

func TestWhenAppendingConversationMessages(t *testing.T) {
	t.Parallel()

	now := time.Now()
	conversationId := "conv-1"

	t.Run("Given an empty conversationId", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {}

		t.Run("Should no-op without touching the repository", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			message := model_builders.NewConversationMessageBuilder().WithTimestamp(now).Build()
			res, err := suite.sut.Append(context.Background(), "   ", []models.ConversationMessage{message})

			assert.NoError(t, err)
			assert.Empty(t, res)
		})
	})

	t.Run("Given an empty messages slice", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {}

		t.Run("Should no-op", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			res, err := suite.sut.Append(context.Background(), conversationId, nil)

			assert.NoError(t, err)
			assert.Empty(t, res)
		})
	})

	t.Run("Given inputs whose hashes are all new", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.expectAgentResolutionAndEnsure()

			suite.mockRepository.EXPECT().ListExistingHashes(gomock.Any(),
				gomock.Any()).
				Return([]string{}, nil)

			suite.mockRepository.EXPECT().AppendMessages(gomock.Any(),
				gomock.Any()).
				DoAndReturn(func(_ context.Context, messages []models.ConversationMessage) ([]models.ConversationMessage, error) {
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
			res, err := suite.sut.Append(context.Background(), conversationId, input)

			assert.NoError(t, err)
			assert.Len(t, res, 2)
			// First input had unicode emoji — content is demojized in the persisted row.
			assert.Equal(t, "hi :fire:", res[0].Content)
		})
	})

	t.Run("Given some inputs match existing hashes", func(t *testing.T) {
		t.Parallel()

		duplicateInput := model_builders.NewConversationMessageBuilder().WithConversationId(conversationId).WithRole("user").WithContent("dup").WithTimestamp(now).Build()
		dupHash, _ := utils.ComputeConversationMessageHash(duplicateInput.Role, duplicateInput.Content, conversationId)

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.expectAgentResolutionAndEnsure()

			suite.mockRepository.EXPECT().ListExistingHashes(gomock.Any(),
				gomock.Any()).
				Return([]string{dupHash}, nil)

			suite.mockRepository.EXPECT().AppendMessages(gomock.Any(),
				gomock.Any()).
				DoAndReturn(func(_ context.Context, messages []models.ConversationMessage) ([]models.ConversationMessage, error) {
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
			res, err := suite.sut.Append(context.Background(), conversationId, input)

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
			suite.expectAgentResolutionAndEnsure()

			suite.mockRepository.EXPECT().ListExistingHashes(gomock.Any(),
				gomock.Any()).
				Return([]string{dupHash}, nil)
			// no AppendMessages call expected
		}

		t.Run("Should short-circuit without inserting", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			res, err := suite.sut.Append(context.Background(), conversationId, []models.ConversationMessage{input})

			assert.NoError(t, err)
			assert.Empty(t, res)
		})
	})

	t.Run("Given ListExistingHashes fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.expectAgentResolutionAndEnsure()

			suite.mockRepository.EXPECT().ListExistingHashes(gomock.Any(),
				gomock.Any()).
				Return(nil, errors.New("db blew"))
		}

		t.Run("Should propagate the error before inserting", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			input := model_builders.NewConversationMessageBuilder().WithTimestamp(now).Build()
			_, err := suite.sut.Append(context.Background(), conversationId, []models.ConversationMessage{input})

			assert.ErrorContains(t, err, "db blew")
		})
	})

	t.Run("Given AppendMessages fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.expectAgentResolutionAndEnsure()

			suite.mockRepository.EXPECT().ListExistingHashes(gomock.Any(),
				gomock.Any()).
				Return([]string{}, nil)
			suite.mockRepository.EXPECT().AppendMessages(gomock.Any(),
				gomock.Any()).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			message := model_builders.NewConversationMessageBuilder().WithTimestamp(now).Build()
			_, err := suite.sut.Append(context.Background(), conversationId, []models.ConversationMessage{message})

			assert.ErrorContains(t, err, "db down")
		})
	})
}
