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
	metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/metrics_mocks/interfaces_mocks"
	knowledge_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	runtime_services_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenAppendingConversationMessagesTestingSuite struct {
	sut *services.MessageAppender

	mockRepository             *knowledge_mocks.MockConversationMessageRepositoryInterface
	mockConversationRepository *knowledge_mocks.MockConversationRepositoryInterface
	mockAgentRepository        *identity_repositories_mocks.MockAgentRepositoryInterface
	mockEmojiDecoder           *runtime_services_mocks.MockEmojiDecoderInterface
}

func WhenAppendingConversationMessagesBeforeEach(t *testing.T) *WhenAppendingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := knowledge_mocks.NewMockConversationMessageRepositoryInterface(mockController)
	mockConversationRepository := knowledge_mocks.NewMockConversationRepositoryInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	mockKnowledgeMetrics := metrics_mocks.NewMockKnowledgeMetricsInterface(mockController)
	mockKnowledgeMetrics.EXPECT().RecordConversationStarted(gomock.Any()).AnyTimes()
	mockKnowledgeMetrics.EXPECT().RecordMessageAppended(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockEmojiDecoder := runtime_services_mocks.NewMockEmojiDecoderInterface(mockController)

	sut := services.NewMessageAppender(
		mockRepository,
		mockConversationRepository,
		mockAgentRepository,
		mockKnowledgeMetrics,
		mockEmojiDecoder,
	)

	return &WhenAppendingConversationMessagesTestingSuite{
		sut:                        sut,
		mockRepository:             mockRepository,
		mockConversationRepository: mockConversationRepository,
		mockAgentRepository:        mockAgentRepository,
		mockEmojiDecoder:           mockEmojiDecoder,
	}
}

func (suite *WhenAppendingConversationMessagesTestingSuite) expectAgentResolutionAndEnsure() {
	suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
	suite.mockConversationRepository.EXPECT().Ensure(gomock.Any(), gomock.Any()).Return(false, nil)
}

func TestWhenAppendingConversationMessages(t *testing.T) {
	t.Parallel()

	now := time.Now()
	conversationId := "conv-1"

	t.Run("Given a blank conversation id", func(t *testing.T) {
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

	t.Run("Given the messages slice is empty", func(t *testing.T) {
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

	t.Run("Given a message carries the system role", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {}

		t.Run("Should reject the message", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			message := model_builders.NewConversationMessageBuilder().
				WithRole("system").
				WithTimestamp(now).
				Build()
			res, err := suite.sut.Append(context.Background(), conversationId, []models.ConversationMessage{message})

			assert.Nil(t, res)
			assert.ErrorContains(t, err, "system messages must never be persisted")
		})
	})

	t.Run("Given listing the existing hashes fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.expectAgentResolutionAndEnsure()

			suite.mockEmojiDecoder.EXPECT().
				DemojizeContent(gomock.Any()).
				DoAndReturn(func(content string) string { return content })

			suite.mockRepository.EXPECT().ListExistingHashes(gomock.Any(),
				gomock.Any()).
				Return(nil, errors.New("db blew"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			input := model_builders.NewConversationMessageBuilder().WithTimestamp(now).Build()
			_, err := suite.sut.Append(context.Background(), conversationId, []models.ConversationMessage{input})

			assert.ErrorContains(t, err, "db blew")
		})
	})

	t.Run("Given the message appender fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.expectAgentResolutionAndEnsure()

			suite.mockEmojiDecoder.EXPECT().
				DemojizeContent(gomock.Any()).
				DoAndReturn(func(content string) string { return content })

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

	t.Run("Given every input is already stored", func(t *testing.T) {
		t.Parallel()

		input := model_builders.NewConversationMessageBuilder().
			WithRole("user").
			WithContent("dup").
			WithTimestamp(now).
			Build()
		input.ConversationId = conversationId
		dupHash, _ := utils.ComputeConversationMessageHash(
			input.Role.String(),
			lo.FromPtr(input.Content),
			"",
			"",
			conversationId,
		)

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.expectAgentResolutionAndEnsure()

			suite.mockEmojiDecoder.EXPECT().
				DemojizeContent(gomock.Any()).
				DoAndReturn(func(content string) string { return content })

			suite.mockRepository.EXPECT().ListExistingHashes(gomock.Any(),
				gomock.Any()).
				Return([]string{dupHash}, nil)
			// no AppendMessages call expected
		}

		t.Run("Should insert nothing", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			res, err := suite.sut.Append(context.Background(), conversationId, []models.ConversationMessage{input})

			assert.NoError(t, err)
			assert.Empty(t, res)
		})
	})

	t.Run("Given inputs whose hashes are all new", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite, persisted *[]models.ConversationMessage) {
			suite.expectAgentResolutionAndEnsure()

			suite.mockEmojiDecoder.EXPECT().DemojizeContent("hi 🔥").Return("hi :fire:")
			suite.mockEmojiDecoder.EXPECT().DemojizeContent("hello").Return("hello")

			suite.mockRepository.EXPECT().ListExistingHashes(gomock.Any(),
				gomock.Any()).
				Return([]string{}, nil)

			suite.mockRepository.EXPECT().AppendMessages(gomock.Any(),
				gomock.Any()).
				DoAndReturn(func(_ context.Context, messages []models.ConversationMessage) ([]models.ConversationMessage, error) {
					*persisted = messages
					return messages, nil
				})
		}

		t.Run("Should return every message", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			var persisted []models.ConversationMessage
			initSuite(suite, &persisted)

			input := []models.ConversationMessage{
				model_builders.NewConversationMessageBuilder().
					WithRole("user").
					WithContent("hi 🔥").
					WithTimestamp(now).
					Build(),
				model_builders.NewConversationMessageBuilder().
					WithRole("assistant").
					WithContent("hello").
					WithTimestamp(now).
					Build(),
			}
			res, err := suite.sut.Append(context.Background(), conversationId, input)

			assert.NoError(t, err)
			assert.Len(t, res, 2)
		})

		t.Run("Should demojize the content", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			var persisted []models.ConversationMessage
			initSuite(suite, &persisted)

			input := []models.ConversationMessage{
				model_builders.NewConversationMessageBuilder().
					WithRole("user").
					WithContent("hi 🔥").
					WithTimestamp(now).
					Build(),
				model_builders.NewConversationMessageBuilder().
					WithRole("assistant").
					WithContent("hello").
					WithTimestamp(now).
					Build(),
			}
			res, err := suite.sut.Append(context.Background(), conversationId, input)

			assert.NoError(t, err)
			assert.Equal(t, "hi :fire:", lo.FromPtr(res[0].Content))
		})

		t.Run("Should stamp the conversation id on every row", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			var persisted []models.ConversationMessage
			initSuite(suite, &persisted)

			input := []models.ConversationMessage{
				model_builders.NewConversationMessageBuilder().
					WithRole("user").
					WithContent("hi 🔥").
					WithTimestamp(now).
					Build(),
				model_builders.NewConversationMessageBuilder().
					WithRole("assistant").
					WithContent("hello").
					WithTimestamp(now).
					Build(),
			}
			_, err := suite.sut.Append(context.Background(), conversationId, input)

			assert.NoError(t, err)
			assert.Len(t, persisted, 2)
			for _, m := range persisted {
				assert.Equal(t, conversationId, m.ConversationId)
			}
		})

		t.Run("Should hash every row", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			var persisted []models.ConversationMessage
			initSuite(suite, &persisted)

			input := []models.ConversationMessage{
				model_builders.NewConversationMessageBuilder().
					WithRole("user").
					WithContent("hi 🔥").
					WithTimestamp(now).
					Build(),
				model_builders.NewConversationMessageBuilder().
					WithRole("assistant").
					WithContent("hello").
					WithTimestamp(now).
					Build(),
			}
			_, err := suite.sut.Append(context.Background(), conversationId, input)

			assert.NoError(t, err)
			assert.Len(t, persisted, 2)
			for _, m := range persisted {
				assert.NotEmpty(t, m.Hash)
			}
		})
	})

	t.Run("Given some inputs match existing hashes", func(t *testing.T) {
		t.Parallel()

		duplicateInput := model_builders.NewConversationMessageBuilder().
			WithConversationId(conversationId).
			WithRole("user").
			WithContent("dup").
			WithTimestamp(now).
			Build()
		dupHash, _ := utils.ComputeConversationMessageHash(
			duplicateInput.Role.String(),
			lo.FromPtr(duplicateInput.Content),
			"",
			"",
			conversationId,
		)

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.expectAgentResolutionAndEnsure()

			suite.mockEmojiDecoder.EXPECT().
				DemojizeContent(gomock.Any()).
				DoAndReturn(func(content string) string { return content }).
				Times(2)

			suite.mockRepository.EXPECT().ListExistingHashes(gomock.Any(),
				gomock.Any()).
				Return([]string{dupHash}, nil)

			suite.mockRepository.EXPECT().AppendMessages(gomock.Any(),
				gomock.Any()).
				DoAndReturn(func(_ context.Context, messages []models.ConversationMessage) ([]models.ConversationMessage, error) {
					if assert.Len(t, messages, 1) {
						assert.Equal(t, "fresh", lo.FromPtr(messages[0].Content))
					}
					return messages, nil
				})
		}

		t.Run("Should insert only the fresh rows", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			input := []models.ConversationMessage{
				model_builders.NewConversationMessageBuilder().
					WithRole("user").
					WithContent("dup").
					WithTimestamp(now).
					Build(),
				model_builders.NewConversationMessageBuilder().
					WithRole("user").
					WithContent("fresh").
					WithTimestamp(now).
					Build(),
			}
			res, err := suite.sut.Append(context.Background(), conversationId, input)

			assert.NoError(t, err)
			assert.Len(t, res, 1)
		})
	})
}
