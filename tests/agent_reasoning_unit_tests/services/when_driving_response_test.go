package services_test

import (
	"context"
	"errors"
	"testing"

	knowledge_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/utils"
	knowledge_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenDrivingResponseTestingSuite struct {
	sut *services.ResponseDriver

	mockInferenceLoop              *interfaces_mocks.MockInferenceLoopInterface
	mockMCPRepository              *repositories_mocks.MockMCPConnectionRepositoryInterface
	mockMCPHeaderAuthSecretStorage *interfaces_mocks.MockMCPHeaderAuthSecretStorageInterface
	mockConversationHistory        *knowledge_interfaces_mocks.MockConversationHistoryInterface
}

func WhenDrivingResponseBeforeEach(t *testing.T) *WhenDrivingResponseTestingSuite {
	mockController := gomock.NewController(t)

	mockInferenceLoop := interfaces_mocks.NewMockInferenceLoopInterface(mockController)
	mockMCPRepository := repositories_mocks.NewMockMCPConnectionRepositoryInterface(mockController)
	mockMCPHeaderAuthSecretStorage := interfaces_mocks.NewMockMCPHeaderAuthSecretStorageInterface(mockController)
	mockConversationHistory := knowledge_interfaces_mocks.NewMockConversationHistoryInterface(mockController)

	sut := services.NewResponseDriver(
		zap.NewNop(),
		mockInferenceLoop,
		mockMCPRepository,
		mockMCPHeaderAuthSecretStorage,
		mockConversationHistory,
	)

	return &WhenDrivingResponseTestingSuite{
		sut:                            sut,
		mockInferenceLoop:              mockInferenceLoop,
		mockMCPRepository:              mockMCPRepository,
		mockMCPHeaderAuthSecretStorage: mockMCPHeaderAuthSecretStorage,
		mockConversationHistory:        mockConversationHistory,
	}
}

func TestWhenDrivingResponse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	conversationId := "conv-1"
	inputs := []knowledge_interfaces.ConversationMessage{
		{Role: "user", Content: "hello"},
	}

	t.Run("Given inputs do not match history", func(t *testing.T) {
		t.Parallel()

		expectedFinal := "hi there"

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationHistory.EXPECT().Read(gomock.Any(), conversationId).Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockMCPRepository.EXPECT().List(gomock.Any()).Return([]models.MCPConnection{}, nil)

			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ []*models.MCPConnection, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					if assert.Len(t, msgs, 1) {
						assert.Equal(t, "user", msgs[0].Role)
						assert.Equal(t, "hello", msgs[0].Content)
					}
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: expectedFinal}), nil
				})

			suite.mockConversationHistory.EXPECT().Write(gomock.Any(),
				conversationId, gomock.Any()).
				DoAndReturn(func(_ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					if assert.Len(t, msgs, 2) {
						assert.Equal(t, "user", msgs[0].Role)
						assert.Equal(t, "hello", msgs[0].Content)
						assert.Equal(t, "assistant", msgs[1].Role)
						assert.Equal(t, expectedFinal, msgs[1].Content)
					}
					return msgs, nil
				})
		}

		t.Run("Should persist inputs + final assistant in one Write after the loop", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, inputs)
			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, "assistant", result.Role)
				assert.Equal(t, expectedFinal, result.Content)
			}
		})
	})

	t.Run("Given an input matches an already-stored message", func(t *testing.T) {
		t.Parallel()

		stored := knowledge_interfaces.ConversationMessage{Role: "user", Content: "hello"}
		freshInput := knowledge_interfaces.ConversationMessage{Role: "user", Content: "follow-up"}

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationHistory.EXPECT().Read(gomock.Any(),
				conversationId).
				Return([]knowledge_interfaces.ConversationMessage{stored}, nil)
			suite.mockMCPRepository.EXPECT().List(gomock.Any()).Return([]models.MCPConnection{}, nil)

			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ []*models.MCPConnection, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					if assert.Len(t, msgs, 2) {
						assert.Equal(t, "hello", msgs[0].Content)
						assert.Equal(t, "follow-up", msgs[1].Content)
					}
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: "ok"}), nil
				})

			suite.mockConversationHistory.EXPECT().Write(gomock.Any(),
				conversationId, gomock.Any()).
				DoAndReturn(func(_ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					if assert.Len(t, msgs, 2) {
						assert.Equal(t, "follow-up", msgs[0].Content)
						assert.Equal(t, "ok", msgs[1].Content)
					}
					return msgs, nil
				})
		}

		t.Run("Should persist only fresh input + final assistant", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, []knowledge_interfaces.ConversationMessage{stored, freshInput})
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given every input is already in history", func(t *testing.T) {
		t.Parallel()

		stored := knowledge_interfaces.ConversationMessage{Role: "user", Content: "hello"}

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationHistory.EXPECT().Read(gomock.Any(),
				conversationId).
				Return([]knowledge_interfaces.ConversationMessage{stored}, nil)
			suite.mockMCPRepository.EXPECT().List(gomock.Any()).Return([]models.MCPConnection{}, nil)
			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ []*models.MCPConnection, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: "ok"}), nil
				})

			suite.mockConversationHistory.EXPECT().Write(gomock.Any(),
				conversationId, gomock.Any()).
				DoAndReturn(func(_ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					if assert.Len(t, msgs, 1) {
						assert.Equal(t, "assistant", msgs[0].Role)
					}
					return msgs, nil
				})
		}

		t.Run("Should persist only the final assistant message", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, []knowledge_interfaces.ConversationMessage{stored})
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given inputs contain unicode emoji that match a demojized history entry", func(t *testing.T) {
		t.Parallel()

		stored := knowledge_interfaces.ConversationMessage{Role: "user", Content: "hi :fire:"}
		emojiInput := knowledge_interfaces.ConversationMessage{Role: "user", Content: "hi 🔥"}

		// Sanity: the hash of the demojized input matches the stored one.
		storedHash, _ := utils.ComputeConversationMessageHash(stored.Role, stored.Content, conversationId)
		demojized := utils.DemojizeContent(emojiInput.Content)
		inputHash, _ := utils.ComputeConversationMessageHash(emojiInput.Role, demojized, conversationId)
		assert.Equal(t, storedHash, inputHash)

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationHistory.EXPECT().Read(gomock.Any(),
				conversationId).
				Return([]knowledge_interfaces.ConversationMessage{stored}, nil)
			suite.mockMCPRepository.EXPECT().List(gomock.Any()).Return([]models.MCPConnection{}, nil)
			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any()).
				Return([]*interfaces.ChatCompletionMessage{
					{Role: "user", Content: "hi :fire:"},
					{Role: "assistant", Content: "ok"},
				}, nil)
			suite.mockConversationHistory.EXPECT().Write(gomock.Any(),
				conversationId, gomock.Any()).
				DoAndReturn(func(_ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					if assert.Len(t, msgs, 1) {
						assert.Equal(t, "assistant", msgs[0].Role)
					}
					return msgs, nil
				})
		}

		t.Run("Should treat the emoji input as a duplicate", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, []knowledge_interfaces.ConversationMessage{emojiInput})
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given MCP connections need header auth hydration", func(t *testing.T) {
		t.Parallel()

		mcpConnection := *model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			WithHeaderAuthKeys([]string{"Authorization"}).
			Build()
		expectedFinal := "ok"

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationHistory.EXPECT().Read(gomock.Any(), conversationId).Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockMCPRepository.EXPECT().List(gomock.Any()).Return([]models.MCPConnection{mcpConnection}, nil)

			suite.mockMCPHeaderAuthSecretStorage.EXPECT().
				GetValue(ctx, mcpConnection.Id, "Authorization").
				Return("Bearer secret", true, nil)

			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, mcps []*models.MCPConnection, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					if assert.Len(t, mcps, 1) {
						assert.Equal(t, "Bearer secret", mcps[0].HeaderAuth["Authorization"])
					}
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: expectedFinal}), nil
				})

			suite.mockConversationHistory.EXPECT().Write(gomock.Any(),
				conversationId, gomock.Any()).
				DoAndReturn(func(_ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					return msgs, nil
				})
		}

		t.Run("Should pass hydrated connections to the loop", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, inputs)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given the inference loop returns no assistant text", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationHistory.EXPECT().Read(gomock.Any(), conversationId).Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockMCPRepository.EXPECT().List(gomock.Any()).Return([]models.MCPConnection{}, nil)
			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ []*models.MCPConnection, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: ""}), nil
				})

			// Inputs still persist even without a final assistant text.
			suite.mockConversationHistory.EXPECT().Write(gomock.Any(),
				conversationId, gomock.Any()).
				DoAndReturn(func(_ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					if assert.Len(t, msgs, 1) {
						assert.Equal(t, "user", msgs[0].Role)
					}
					return msgs, nil
				})
		}

		t.Run("Should persist inputs only and return nil", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, inputs)
			assert.NoError(t, err)
			assert.Nil(t, result)
		})
	})

	t.Run("Given reading history fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationHistory.EXPECT().Read(gomock.Any(),
				conversationId).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should fail before any Write", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, inputs)
			assert.Nil(t, result)
			assert.Error(t, err)
			assert.ErrorContains(t, err, "failed to load conversation history")
		})
	})

	t.Run("Given the inference loop fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationHistory.EXPECT().Read(gomock.Any(), conversationId).Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockMCPRepository.EXPECT().List(gomock.Any()).Return([]models.MCPConnection{}, nil)

			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any()).
				Return(nil, errors.New("llm broke"))
			// no Write expected — atomic semantics: nothing persists when the loop fails
		}

		t.Run("Should propagate the error without persisting anything", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, inputs)
			assert.Nil(t, result)
			assert.Error(t, err)
			assert.ErrorContains(t, err, "inference loop failed")
		})
	})

	t.Run("Given the post-loop Write fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationHistory.EXPECT().Read(gomock.Any(), conversationId).Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockMCPRepository.EXPECT().List(gomock.Any()).Return([]models.MCPConnection{}, nil)
			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ []*models.MCPConnection, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: "ok"}), nil
				})
			suite.mockConversationHistory.EXPECT().Write(gomock.Any(),
				conversationId, gomock.Any()).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should wrap the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, inputs)
			assert.Nil(t, result)
			assert.Error(t, err)
			assert.ErrorContains(t, err, "failed to persist conversation")
		})
	})

	t.Run("Given the MCP header auth lookup fails", func(t *testing.T) {
		t.Parallel()

		mcpConnection := *model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			WithHeaderAuthKeys([]string{"Authorization"}).
			Build()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationHistory.EXPECT().Read(gomock.Any(), conversationId).Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockMCPRepository.EXPECT().List(gomock.Any()).Return([]models.MCPConnection{mcpConnection}, nil)
			suite.mockMCPHeaderAuthSecretStorage.EXPECT().
				GetValue(ctx, mcpConnection.Id, "Authorization").
				Return("", false, errors.New("vault down"))
		}

		t.Run("Should fail before invoking the loop", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, inputs)
			assert.Nil(t, result)
			assert.Error(t, err)
			assert.ErrorContains(t, err, "failed to load header auth")
		})
	})
}
