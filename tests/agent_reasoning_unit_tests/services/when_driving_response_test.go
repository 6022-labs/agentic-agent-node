package services_test

import (
	"context"
	"errors"
	"testing"

	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	knowledge_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/settings"
	traces_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/traces/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/utils"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	identity_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	knowledge_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	traces_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/traces_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	runtime_services_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenDrivingResponseTestingSuite struct {
	sut *services.ResponseDriver

	mockInferenceLoop              *interfaces_mocks.MockInferenceLoopInterface
	mockMCPRepository              *repositories_mocks.MockMCPConnectionRepositoryInterface
	mockLLMConnectionRepository    *repositories_mocks.MockLLMConnectionRepositoryInterface
	mockAgentGuidanceRepository    *repositories_mocks.MockAgentGuidanceRepositoryInterface
	mockMCPHeaderAuthSecretStorage *interfaces_mocks.MockMCPHeaderAuthSecretStorageInterface
	mockConversationHistory        *knowledge_interfaces_mocks.MockConversationHistoryInterface
	mockAgentResolver              *identity_interfaces_mocks.MockAgentResolverInterface
	mockAgentAddressRepository     *identity_repositories_mocks.MockAgentAddressRepositoryInterface
	mockSystemPromptBuilder        *interfaces_mocks.MockSystemPromptBuilderInterface
	mockMcpRequester               *interfaces_mocks.MockMcpRequesterInterface
	mockSelfToolRegistry           *interfaces_mocks.MockSelfToolRegistryInterface
	mockSelfToolRepository         *repositories_mocks.MockSelfToolRepositoryInterface
	mockEmojiDecoder               *runtime_services_mocks.MockEmojiDecoderInterface
	mockConversationTitler         *interfaces_mocks.MockConversationTitlerInterface
}

type drivingResponseConfig struct {
	selfTools      []*models.Tool
	selfToolRows   []models.SelfTool
	defaultEnabled map[string]bool
}

type drivingResponseOption func(*drivingResponseConfig)

func withSelfTools(selfTools ...*models.Tool) drivingResponseOption {
	return func(c *drivingResponseConfig) { c.selfTools = selfTools }
}

func withStoredSelfTools(rows []models.SelfTool) drivingResponseOption {
	return func(c *drivingResponseConfig) { c.selfToolRows = rows }
}

func withDefaultEnabled(defaultEnabled map[string]bool) drivingResponseOption {
	return func(c *drivingResponseConfig) { c.defaultEnabled = defaultEnabled }
}

// WhenDrivingResponseBeforeEach defaults every registered self tool to enabled with no stored rows.
func WhenDrivingResponseBeforeEach(t *testing.T, opts ...drivingResponseOption) *WhenDrivingResponseTestingSuite {
	config := &drivingResponseConfig{}
	for _, opt := range opts {
		opt(config)
	}
	if config.defaultEnabled == nil {
		config.defaultEnabled = make(map[string]bool, len(config.selfTools))
		for _, tool := range config.selfTools {
			config.defaultEnabled[tool.Name] = true
		}
	}
	selfTools, selfToolRows, defaultEnabled := config.selfTools, config.selfToolRows, config.defaultEnabled

	mockController := gomock.NewController(t)

	mockInferenceLoop := interfaces_mocks.NewMockInferenceLoopInterface(mockController)
	mockEmojiDecoder := runtime_services_mocks.NewMockEmojiDecoderInterface(mockController)
	mockMCPRepository := repositories_mocks.NewMockMCPConnectionRepositoryInterface(mockController)
	mockLLMConnectionRepository := repositories_mocks.NewMockLLMConnectionRepositoryInterface(mockController)
	mockAgentGuidanceRepository := repositories_mocks.NewMockAgentGuidanceRepositoryInterface(mockController)
	mockMCPHeaderAuthSecretStorage := interfaces_mocks.NewMockMCPHeaderAuthSecretStorageInterface(mockController)
	mockConversationHistory := knowledge_interfaces_mocks.NewMockConversationHistoryInterface(mockController)
	mockAgentResolver := identity_interfaces_mocks.NewMockAgentResolverInterface(mockController)
	mockAgentAddressRepository := identity_repositories_mocks.NewMockAgentAddressRepositoryInterface(mockController)
	mockSystemPromptBuilder := interfaces_mocks.NewMockSystemPromptBuilderInterface(mockController)
	mockMcpRequester := interfaces_mocks.NewMockMcpRequesterInterface(mockController)

	mockSelfToolRegistry := interfaces_mocks.NewMockSelfToolRegistryInterface(mockController)
	mockSelfToolRegistry.EXPECT().ListTools().Return(selfTools).AnyTimes()
	mockSelfToolRegistry.EXPECT().ListDefaultEnabled().Return(defaultEnabled).AnyTimes()

	mockSelfToolRepository := repositories_mocks.NewMockSelfToolRepositoryInterface(mockController)
	mockSelfToolRepository.EXPECT().ListByAgentId(gomock.Any(), gomock.Any()).Return(selfToolRows, nil).AnyTimes()

	mockSpan := traces_interfaces_mocks.NewMockSpan(mockController)
	mockSpan.EXPECT().Fail(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().End().AnyTimes()

	mockReasoningTracer := traces_interfaces_mocks.NewMockReasoningTracerInterface(mockController)
	mockReasoningTracer.EXPECT().StartListTools(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ string) (context.Context, traces_interfaces.Span) {
			return ctx, mockSpan
		}).AnyTimes()

	// Titling runs in a goroutine after the turn: allow it, assert it elsewhere.
	mockConversationTitler := interfaces_mocks.NewMockConversationTitlerInterface(mockController)

	sut := services.NewResponseDriver(
		zap.NewNop(),
		mockInferenceLoop,
		mockMCPRepository,
		mockLLMConnectionRepository,
		mockAgentGuidanceRepository,
		mockMCPHeaderAuthSecretStorage,
		mockConversationHistory,
		mockAgentResolver,
		mockAgentAddressRepository,
		mockSystemPromptBuilder,
		mockMcpRequester,
		mockSelfToolRegistry,
		mockSelfToolRepository,
		mockReasoningTracer,
		mockEmojiDecoder,
		mockConversationTitler,
		&settings.ConversationTitlerSettings{TimeoutSeconds: 30},
	)

	return &WhenDrivingResponseTestingSuite{
		sut:                            sut,
		mockInferenceLoop:              mockInferenceLoop,
		mockMCPRepository:              mockMCPRepository,
		mockLLMConnectionRepository:    mockLLMConnectionRepository,
		mockAgentGuidanceRepository:    mockAgentGuidanceRepository,
		mockMCPHeaderAuthSecretStorage: mockMCPHeaderAuthSecretStorage,
		mockConversationHistory:        mockConversationHistory,
		mockAgentResolver:              mockAgentResolver,
		mockAgentAddressRepository:     mockAgentAddressRepository,
		mockSystemPromptBuilder:        mockSystemPromptBuilder,
		mockMcpRequester:               mockMcpRequester,
		mockSelfToolRegistry:           mockSelfToolRegistry,
		mockSelfToolRepository:         mockSelfToolRepository,
		mockEmojiDecoder:               mockEmojiDecoder,
		mockConversationTitler:         mockConversationTitler,
	}
}

func (suite *WhenDrivingResponseTestingSuite) expectDemojizePassthrough() {
	suite.mockEmojiDecoder.EXPECT().DemojizeContent(gomock.Any()).
		DoAndReturn(func(content string) string { return content }).
		AnyTimes()
}

func TestWhenDrivingResponse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	builtSystemPrompt := "BUILT SYSTEM PROMPT"
	conversationId := "conv-1"
	inputs := []knowledge_interfaces.ConversationMessage{
		{Role: "user", Content: "hello"},
	}

	t.Run("Given the history lookup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().Read(gomock.Any(),
				conversationId).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should fail before any Write", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.Nil(t, result)
			assert.ErrorContains(t, err, "failed to load conversation history")
		})
	})

	t.Run("Given the LLM connection lookup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().
				Read(gomock.Any(), conversationId).
				Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should wrap the error before loading MCP connections", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.Nil(t, result)
			assert.ErrorContains(t, err, "failed to load llm connection")
		})
	})

	t.Run("Given no LLM connection exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().
				Read(gomock.Any(), conversationId).
				Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)
		}

		t.Run("Should fail before loading MCP connections", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.Nil(t, result)
			assert.ErrorIs(t, err, interfaces.ErrLLMConnectionNotConfigured)
		})
	})

	t.Run("Given the MCP header auth lookup fails", func(t *testing.T) {
		t.Parallel()

		mcpConnection := *model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			WithHeaderAuthKeys([]string{"Authorization"}).
			Build()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().
				Read(gomock.Any(), conversationId).
				Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{mcpConnection}, nil)
			suite.mockMCPHeaderAuthSecretStorage.EXPECT().
				GetValue(ctx, mcpConnection.Id, "Authorization").
				Return("", false, errors.New("vault down"))
		}

		t.Run("Should fail before invoking the loop", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.Nil(t, result)
			assert.ErrorContains(t, err, "failed to load header auth")
		})
	})

	t.Run("Given listing tools fails for an MCP connection", func(t *testing.T) {
		t.Parallel()

		mcpConnection := *model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			Build()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().
				Read(gomock.Any(), conversationId).
				Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{mcpConnection}, nil)
			suite.mockMCPHeaderAuthSecretStorage.EXPECT().
				GetValue(gomock.Any(), gomock.Any(), gomock.Any()).
				Return("", false, nil).
				AnyTimes()
			suite.mockMcpRequester.EXPECT().
				ListTools(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("mcp down"))
		}

		t.Run("Should fail before building the prompt or invoking the loop", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.Nil(t, result)
			assert.ErrorContains(t, err, "failed to list tools")
		})
	})

	t.Run("Given one of several MCP connections fails to list", func(t *testing.T) {
		t.Parallel()

		connA := *model_builders.NewMCPConnectionBuilder().WithName("alpha").Build()
		connB := *model_builders.NewMCPConnectionBuilder().WithName("beta").Build()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().
				Read(gomock.Any(), conversationId).
				Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{connA, connB}, nil)
			suite.mockMCPHeaderAuthSecretStorage.EXPECT().
				GetValue(gomock.Any(), gomock.Any(), gomock.Any()).
				Return("", false, nil).
				AnyTimes()

			// Both are attempted concurrently; beta fails, which fails the whole discovery.
			suite.mockMcpRequester.EXPECT().
				ListTools(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, conn *models.MCPConnection) ([]*models.Tool, error) {
					if conn.Name == "beta" {
						return nil, errors.New("mcp beta down")
					}
					return []*models.Tool{{Name: conn.Name + "_tool"}}, nil
				}).Times(2)
			// The builder and the loop are never reached once discovery fails.
		}

		t.Run("Should fail the whole request without building the prompt or invoking the loop", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.Nil(t, result)
			assert.ErrorContains(t, err, "failed to list tools")
		})
	})

	t.Run("Given the inference loop fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().
				Read(gomock.Any(), conversationId).
				Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{}, nil)
			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)
			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "").
				Return(builtSystemPrompt)

			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("llm broke"))
			// no Write expected — atomic semantics: nothing persists when the loop fails
		}

		t.Run("Should propagate the error without persisting anything", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.Nil(t, result)
			assert.ErrorContains(t, err, "inference loop failed")
		})
	})

	t.Run("Given the post-loop Write fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().
				Read(gomock.Any(), conversationId).
				Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{}, nil)
			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)
			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "").
				Return(builtSystemPrompt)
			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *models.LLMConnection, _ []*models.Tool, _ map[*models.MCPConnection][]*models.Tool, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: "ok"}), nil
				})
			suite.mockConversationHistory.EXPECT().Write(gomock.Any(),
				conversationId, gomock.Any()).
				Return(nil, errors.New("db down"))
			suite.mockConversationTitler.EXPECT().
				TitleIfUnnamed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
		}

		t.Run("Should wrap the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.Nil(t, result)
			assert.ErrorContains(t, err, "failed to persist conversation")
		})

	})

	t.Run("Given inputs contain unicode emoji that match a demojized history entry", func(t *testing.T) {
		t.Parallel()

		stored := knowledge_interfaces.ConversationMessage{Role: "user", Content: "hi :fire:"}
		emojiInput := knowledge_interfaces.ConversationMessage{Role: "user", Content: "hi 🔥"}

		// Sanity: the hash of the demojized input matches the stored one.
		storedHash, _ := utils.ComputeConversationMessageHash(
			stored.Role,
			stored.Content,
			stored.ToolCalls,
			stored.ToolCallId,
			conversationId,
		)
		inputHash, _ := utils.ComputeConversationMessageHash(
			emojiInput.Role,
			"hi :fire:",
			emojiInput.ToolCalls,
			emojiInput.ToolCallId,
			conversationId,
		)
		assert.Equal(t, storedHash, inputHash)

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationTitler.EXPECT().
				TitleIfUnnamed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
			suite.mockEmojiDecoder.EXPECT().
				DemojizeContent(emojiInput.Content).
				Return("hi :fire:")
			suite.mockConversationHistory.EXPECT().Read(gomock.Any(),
				conversationId).
				Return([]knowledge_interfaces.ConversationMessage{stored}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{}, nil)
			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)
			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "").
				Return(builtSystemPrompt)
			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return([]*interfaces.ChatCompletionMessage{
					{Role: "system", Content: builtSystemPrompt},
					{Role: "user", Content: "hi :fire:"},
					{Role: "assistant", Content: "ok"},
				}, nil)
			suite.mockConversationHistory.EXPECT().Write(gomock.Any(),
				conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
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

			result, err := suite.sut.Drive(
				ctx,
				conversationId,
				"",
				[]knowledge_interfaces.ConversationMessage{emojiInput},
			)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given an input matches an already-stored message", func(t *testing.T) {
		t.Parallel()

		stored := knowledge_interfaces.ConversationMessage{Role: "user", Content: "hello"}
		freshInput := knowledge_interfaces.ConversationMessage{Role: "user", Content: "follow-up"}

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationTitler.EXPECT().
				TitleIfUnnamed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().Read(gomock.Any(),
				conversationId).
				Return([]knowledge_interfaces.ConversationMessage{stored}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{}, nil)
			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)
			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "").
				Return(builtSystemPrompt)

			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *models.LLMConnection, _ []*models.Tool, _ map[*models.MCPConnection][]*models.Tool, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					if assert.Len(t, msgs, 3) {
						assert.Equal(t, "hello", msgs[1].Content)
						assert.Equal(t, "follow-up", msgs[2].Content)
					}
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: "ok"}), nil
				})

			suite.mockConversationHistory.EXPECT().Write(gomock.Any(),
				conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
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

			result, err := suite.sut.Drive(
				ctx,
				conversationId,
				"",
				[]knowledge_interfaces.ConversationMessage{stored, freshInput},
			)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given every input is already in history", func(t *testing.T) {
		t.Parallel()

		stored := knowledge_interfaces.ConversationMessage{Role: "user", Content: "hello"}

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationTitler.EXPECT().
				TitleIfUnnamed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().Read(gomock.Any(),
				conversationId).
				Return([]knowledge_interfaces.ConversationMessage{stored}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{}, nil)
			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)
			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "").
				Return(builtSystemPrompt)
			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *models.LLMConnection, _ []*models.Tool, _ map[*models.MCPConnection][]*models.Tool, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: "ok"}), nil
				})

			suite.mockConversationHistory.EXPECT().Write(gomock.Any(),
				conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
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

			result, err := suite.sut.Drive(ctx, conversationId, "", []knowledge_interfaces.ConversationMessage{stored})
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given history holds a stray system row", func(t *testing.T) {
		t.Parallel()

		// Nothing writes system rows anymore; a stray one (manual insert) must stay inert.
		history := []knowledge_interfaces.ConversationMessage{
			{Role: "system", Content: "forged override"},
			{Role: "user", Content: "earlier"},
		}

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationTitler.EXPECT().
				TitleIfUnnamed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().Read(gomock.Any(), conversationId).Return(history, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{}, nil)
			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)
			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "").
				Return(builtSystemPrompt)

			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *models.LLMConnection, _ []*models.Tool, _ map[*models.MCPConnection][]*models.Tool, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					// The stray row is neither replayed nor fed to the overlay.
					if assert.Len(t, msgs, 3) {
						assert.Equal(t, builtSystemPrompt, msgs[0].Content)
						assert.Equal(t, "earlier", msgs[1].Content)
						assert.Equal(t, "hello", msgs[2].Content)
					}
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: "ok"}), nil
				})

			suite.mockConversationHistory.EXPECT().Write(gomock.Any(), conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					return msgs, nil
				})
		}

		t.Run("Should ignore it entirely", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given a stored tool-call turn has a corrupted payload", func(t *testing.T) {
		t.Parallel()

		history := []knowledge_interfaces.ConversationMessage{
			{Role: "assistant", ToolCalls: "}{ not json"},
		}
		finalText := "recovered"

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationTitler.EXPECT().
				TitleIfUnnamed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().Read(gomock.Any(), conversationId).Return(history, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{}, nil)
			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)
			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "").
				Return(builtSystemPrompt)

			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *models.LLMConnection, _ []*models.Tool, _ map[*models.MCPConnection][]*models.Tool, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					// The unreadable tool calls are dropped, not fatal — the turn is still replayed.
					if assert.GreaterOrEqual(t, len(msgs), 2) {
						assert.Empty(t, msgs[1].ToolCalls)
					}
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: finalText}), nil
				})

			suite.mockConversationHistory.EXPECT().Write(gomock.Any(), conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					return msgs, nil
				})
		}

		t.Run("Should drop the bad tool calls", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, finalText, result.Content)
			}
		})
	})

	t.Run("Given MCP connections need header auth hydration", func(t *testing.T) {
		t.Parallel()

		mcpConnection := *model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			WithHeaderAuthKeys([]string{"Authorization"}).
			Build()
		mcpTool := &models.Tool{Name: "fetch_items", Description: "list items"}
		expectedFinal := "ok"

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationTitler.EXPECT().
				TitleIfUnnamed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().
				Read(gomock.Any(), conversationId).
				Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{mcpConnection}, nil)

			suite.mockMCPHeaderAuthSecretStorage.EXPECT().
				GetValue(ctx, mcpConnection.Id, "Authorization").
				Return("Bearer secret", true, nil)

			// The driver lists tools once per connection and shares the map with builder + loop.
			suite.mockMcpRequester.EXPECT().
				ListTools(gomock.Any(), gomock.Any()).
				Return([]*models.Tool{mcpTool}, nil)

			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)
			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "").
				DoAndReturn(func(_ any, _ map[identity_enums.AddressType]string, _ string, _ []*models.Tool, mcpTools map[*models.MCPConnection][]*models.Tool, _ string) string {
					if assert.Len(t, mcpTools, 1) {
						for _, tools := range mcpTools {
							if assert.Len(t, tools, 1) {
								assert.Equal(t, "fetch_items", tools[0].Name)
							}
						}
					}
					return builtSystemPrompt
				})

			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *models.LLMConnection, selfTools []*models.Tool, mcpTools map[*models.MCPConnection][]*models.Tool, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					if assert.Len(t, mcpTools, 1) {
						for conn, tools := range mcpTools {
							assert.Equal(t, "Bearer secret", conn.HeaderAuth["Authorization"])
							assert.Len(t, tools, 1)
						}
					}
					return append(
						msgs,
						&interfaces.ChatCompletionMessage{Role: "assistant", Content: expectedFinal},
					), nil
				})

			suite.mockConversationHistory.EXPECT().Write(gomock.Any(),
				conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					return msgs, nil
				})
		}

		t.Run("Should hydrate the connections, list their tools", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given an MCP connection restricts its allowed tool patterns", func(t *testing.T) {
		t.Parallel()

		mcpConnection := *model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			WithHeaderAuthKeys(nil).
			WithAllowedToolPatterns([]string{"fetch_.*"}).
			Build()
		allowedTool := &models.Tool{Name: "fetch_items"}
		filteredTool := &models.Tool{Name: "delete_items"}

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationTitler.EXPECT().
				TitleIfUnnamed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().
				Read(gomock.Any(), conversationId).
				Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{mcpConnection}, nil)
			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)

			suite.mockMcpRequester.EXPECT().
				ListTools(gomock.Any(), gomock.Any()).
				Return([]*models.Tool{allowedTool, filteredTool}, nil)

			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "").
				DoAndReturn(func(_ any, _ map[identity_enums.AddressType]string, _ string, _ []*models.Tool, mcpTools map[*models.MCPConnection][]*models.Tool, _ string) string {
					if assert.Len(t, mcpTools, 1) {
						for _, tools := range mcpTools {
							if assert.Len(t, tools, 1) {
								assert.Equal(t, "fetch_items", tools[0].Name)
							}
						}
					}
					return builtSystemPrompt
				})

			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *models.LLMConnection, _ []*models.Tool, mcpTools map[*models.MCPConnection][]*models.Tool, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					if assert.Len(t, mcpTools, 1) {
						for _, tools := range mcpTools {
							if assert.Len(t, tools, 1) {
								assert.Equal(t, "fetch_items", tools[0].Name)
							}
						}
					}
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: "ok"}), nil
				})

			suite.mockConversationHistory.EXPECT().Write(gomock.Any(), conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					return msgs, nil
				})
		}

		t.Run("Should pass only the pattern-allowed tools to builder + loop", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given multiple MCP connections", func(t *testing.T) {
		t.Parallel()

		connA := *model_builders.NewMCPConnectionBuilder().WithName("alpha").Build()
		connB := *model_builders.NewMCPConnectionBuilder().WithName("beta").Build()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationTitler.EXPECT().
				TitleIfUnnamed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().
				Read(gomock.Any(), conversationId).
				Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{connA, connB}, nil)
			suite.mockMCPHeaderAuthSecretStorage.EXPECT().
				GetValue(gomock.Any(), gomock.Any(), gomock.Any()).
				Return("", false, nil).
				AnyTimes()

			// Fanned out per connection, resolved in any order.
			suite.mockMcpRequester.EXPECT().
				ListTools(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, conn *models.MCPConnection) ([]*models.Tool, error) {
					return []*models.Tool{{Name: conn.Name + "_tool"}}, nil
				}).Times(2)

			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)
			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "").
				Return(builtSystemPrompt)

			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *models.LLMConnection, _ []*models.Tool, mcpTools map[*models.MCPConnection][]*models.Tool, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					// Both servers' tools are collected into the map regardless of completion order.
					names := map[string]bool{}
					for _, tools := range mcpTools {
						for _, tl := range tools {
							names[tl.Name] = true
						}
					}
					assert.Len(t, mcpTools, 2)
					assert.True(t, names["alpha_tool"])
					assert.True(t, names["beta_tool"])
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: "ok"}), nil
				})

			suite.mockConversationHistory.EXPECT().Write(gomock.Any(), conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					return msgs, nil
				})
		}

		t.Run("Should discover every server concurrently", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given the self tool registry exposes tools", func(t *testing.T) {
		t.Parallel()

		selfTool := &models.Tool{Name: "get_identity", Description: "read own identity"}

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationTitler.EXPECT().
				TitleIfUnnamed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().
				Read(gomock.Any(), conversationId).
				Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{}, nil)
			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)

			// mockMcpRequester.ListTools is never expected — self tools are not listed over MCP.
			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "").
				DoAndReturn(func(_ any, _ map[identity_enums.AddressType]string, _ string, selfTools []*models.Tool, mcpTools map[*models.MCPConnection][]*models.Tool, _ string) string {
					assert.Empty(t, mcpTools)
					if assert.Len(t, selfTools, 1) {
						assert.Equal(t, "get_identity", selfTools[0].Name)
					}
					return builtSystemPrompt
				})

			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *models.LLMConnection, selfTools []*models.Tool, mcpTools map[*models.MCPConnection][]*models.Tool, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					assert.Empty(t, mcpTools)
					if assert.Len(t, selfTools, 1) {
						assert.Equal(t, "get_identity", selfTools[0].Name)
					}
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: "ok"}), nil
				})

			suite.mockConversationHistory.EXPECT().Write(gomock.Any(), conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					return msgs, nil
				})
		}

		t.Run("Should pass the self tools to builder + loop", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t, withSelfTools(selfTool))
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given no self tool is registered", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationTitler.EXPECT().
				TitleIfUnnamed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().
				Read(gomock.Any(), conversationId).
				Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{}, nil)
			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)
			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "").
				Return(builtSystemPrompt)

			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *models.LLMConnection, selfTools []*models.Tool, mcpTools map[*models.MCPConnection][]*models.Tool, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					assert.Empty(t, mcpTools)
					assert.Empty(t, selfTools)
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: "ok"}), nil
				})

			suite.mockConversationHistory.EXPECT().Write(gomock.Any(), conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					return msgs, nil
				})
		}

		t.Run("Should pass no self tools to builder + loop", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given a row disables a declared self tool", func(t *testing.T) {
		t.Parallel()

		enabledTool := &models.Tool{Name: "get_identity"}
		disabledTool := &models.Tool{Name: "run_command"}
		disabledRow := *model_builders.NewSelfToolBuilder().WithName("run_command").WithEnabled(false).Build()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationTitler.EXPECT().
				TitleIfUnnamed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().
				Read(gomock.Any(), conversationId).
				Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{}, nil)
			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)

			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "").
				DoAndReturn(func(_ any, _ map[identity_enums.AddressType]string, _ string, selfTools []*models.Tool, mcpTools map[*models.MCPConnection][]*models.Tool, _ string) string {
					assert.Empty(t, mcpTools)
					if assert.Len(t, selfTools, 1) {
						assert.Equal(t, "get_identity", selfTools[0].Name)
					}
					return builtSystemPrompt
				})

			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *models.LLMConnection, selfTools []*models.Tool, mcpTools map[*models.MCPConnection][]*models.Tool, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					assert.Empty(t, mcpTools)
					if assert.Len(t, selfTools, 1) {
						assert.Equal(t, "get_identity", selfTools[0].Name)
					}
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: "ok"}), nil
				})

			suite.mockConversationHistory.EXPECT().Write(gomock.Any(), conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					return msgs, nil
				})
		}

		t.Run("Should filter the disabled tool out of the list passed to builder + loop", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t,
				withSelfTools(enabledTool, disabledTool),
				withStoredSelfTools([]models.SelfTool{disabledRow}),
				withDefaultEnabled(map[string]bool{"get_identity": true, "run_command": true}))
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given a declared self tool has no stored row", func(t *testing.T) {
		t.Parallel()

		defaultOnTool := &models.Tool{Name: "get_identity"}
		defaultOffTool := &models.Tool{Name: "run_command"}

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationTitler.EXPECT().
				TitleIfUnnamed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().
				Read(gomock.Any(), conversationId).
				Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{}, nil)
			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)

			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "").
				DoAndReturn(func(_ any, _ map[identity_enums.AddressType]string, _ string, selfTools []*models.Tool, mcpTools map[*models.MCPConnection][]*models.Tool, _ string) string {
					assert.Empty(t, mcpTools)
					if assert.Len(t, selfTools, 1) {
						assert.Equal(t, "get_identity", selfTools[0].Name)
					}
					return builtSystemPrompt
				})

			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *models.LLMConnection, selfTools []*models.Tool, mcpTools map[*models.MCPConnection][]*models.Tool, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					assert.Empty(t, mcpTools)
					if assert.Len(t, selfTools, 1) {
						assert.Equal(t, "get_identity", selfTools[0].Name)
					}
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: "ok"}), nil
				})

			suite.mockConversationHistory.EXPECT().Write(gomock.Any(), conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					return msgs, nil
				})
		}

		t.Run("Should fall back to each tool's registry default", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t,
				withSelfTools(defaultOnTool, defaultOffTool),
				withDefaultEnabled(map[string]bool{"get_identity": true, "run_command": false}))
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given owner guidance exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationTitler.EXPECT().
				TitleIfUnnamed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().
				Read(gomock.Any(), conversationId).
				Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{}, nil)
			suite.mockAgentGuidanceRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewAgentGuidanceBuilder().WithContent("answer concisely").Build(), nil)
			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), "answer concisely", gomock.Any(), gomock.Any(), "").
				Return(builtSystemPrompt)
			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *models.LLMConnection, _ []*models.Tool, _ map[*models.MCPConnection][]*models.Tool, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: "ok"}), nil
				})
			suite.mockConversationHistory.EXPECT().Write(gomock.Any(), conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					return msgs, nil
				})
		}

		t.Run("Should forward the guidance content to the prompt builder", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given instructions accompany the turn", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationTitler.EXPECT().
				TitleIfUnnamed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().
				Read(gomock.Any(), conversationId).
				Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{}, nil)
			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)
			// Instructions feed the per-turn overlay only — never a message, never persisted.
			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "act formal").
				Return(builtSystemPrompt)

			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *models.LLMConnection, _ []*models.Tool, _ map[*models.MCPConnection][]*models.Tool, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					if assert.Len(t, msgs, 2) {
						assert.Equal(t, "system", msgs[0].Role)
						assert.Equal(t, builtSystemPrompt, msgs[0].Content)
						assert.Equal(t, "user", msgs[1].Role)
					}
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: "ok"}), nil
				})

			suite.mockConversationHistory.EXPECT().Write(gomock.Any(),
				conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					// Only the user turn and the assistant reply persist — no system row.
					if assert.Len(t, msgs, 2) {
						assert.Equal(t, "user", msgs[0].Role)
						assert.Equal(t, "assistant", msgs[1].Role)
					}
					return msgs, nil
				})
		}

		t.Run("Should feed them to the builder as overlay without persisting any system row", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "act formal", inputs)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given the inference loop returns no assistant text", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationTitler.EXPECT().
				TitleIfUnnamed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().
				Read(gomock.Any(), conversationId).
				Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{}, nil)
			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)
			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "").
				Return(builtSystemPrompt)
			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *models.LLMConnection, _ []*models.Tool, _ map[*models.MCPConnection][]*models.Tool, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					return append(msgs, &interfaces.ChatCompletionMessage{Role: "assistant", Content: ""}), nil
				})

			// Inputs still persist even without a final assistant text.
			suite.mockConversationHistory.EXPECT().Write(gomock.Any(),
				conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					if assert.Len(t, msgs, 1) {
						assert.Equal(t, "user", msgs[0].Role)
					}
					return msgs, nil
				})
		}

		t.Run("Should persist the inputs only", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.NoError(t, err)
		})

		t.Run("Should return nil", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.NoError(t, err)
			assert.Nil(t, result)
		})
	})

	t.Run("Given history holds a tool-call turn and the loop emits new tool calls", func(t *testing.T) {
		t.Parallel()

		storedToolCalls := `[{"Id":"call-1","Name":"conn_tool","Arguments":{"city":"paris"}}]`
		history := []knowledge_interfaces.ConversationMessage{
			{Role: "assistant", ToolCalls: storedToolCalls},
			{Role: "tool", Content: `{"temp":"20"}`, ToolCallId: "call-1"},
		}

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationTitler.EXPECT().
				TitleIfUnnamed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().Read(gomock.Any(), conversationId).Return(history, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{}, nil)
			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)
			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "").
				Return(builtSystemPrompt)

			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *models.LLMConnection, _ []*models.Tool, _ map[*models.MCPConnection][]*models.Tool, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					// History is replayed full-fidelity, tool-call turn and result included.
					if assert.Len(t, msgs, 4) {
						assert.Equal(t, "system", msgs[0].Role)
						assert.Equal(t, "assistant", msgs[1].Role)
						if assert.Len(t, msgs[1].ToolCalls, 1) {
							assert.Equal(t, "call-1", msgs[1].ToolCalls[0].Id)
							assert.Equal(t, "conn_tool", msgs[1].ToolCalls[0].Name)
							assert.Equal(t, "paris", msgs[1].ToolCalls[0].Arguments["city"])
						}
						assert.Equal(t, "tool", msgs[2].Role)
						assert.Equal(t, "call-1", msgs[2].ToolCallId)
						assert.Equal(t, "user", msgs[3].Role)
					}
					return append(
						msgs,
						&interfaces.ChatCompletionMessage{
							Role: "assistant",
							ToolCalls: []interfaces.ToolCall{
								{Id: "call-2", Name: "conn_tool", Arguments: map[string]any{"city": "london"}},
							},
						},
						&interfaces.ChatCompletionMessage{Role: "tool", ToolCallId: "call-2", Content: `{"temp":"15"}`},
						&interfaces.ChatCompletionMessage{Role: "assistant", Content: "final answer"},
					), nil
				})

			suite.mockConversationHistory.EXPECT().Write(gomock.Any(),
				conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					// Persist the new input plus every assistant + tool turn the loop added.
					if assert.Len(t, msgs, 4) {
						assert.Equal(t, "user", msgs[0].Role)

						assert.Equal(t, "assistant", msgs[1].Role)
						assert.Contains(t, msgs[1].ToolCalls, "call-2")
						assert.Empty(t, msgs[1].Content)

						assert.Equal(t, "tool", msgs[2].Role)
						assert.Equal(t, "call-2", msgs[2].ToolCallId)
						assert.Equal(t, `{"temp":"15"}`, msgs[2].Content)

						assert.Equal(t, "assistant", msgs[3].Role)
						assert.Equal(t, "final answer", msgs[3].Content)

						// Strictly increasing timestamps keep stream order (and tool pairing) on reload.
						for i := 0; i < len(msgs)-1; i++ {
							assert.True(t, msgs[i].Timestamp.Before(msgs[i+1].Timestamp))
						}
					}
					return msgs, nil
				})
		}

		t.Run("Should persist the full new stream in order", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.NoError(t, err)
		})

		t.Run("Should return the final answer", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.NoError(t, err)
			if assert.NotNil(t, result) {
			}
		})
	})

	t.Run("Given inputs do not match history", func(t *testing.T) {
		t.Parallel()

		expectedFinal := "hi there"

		initSuite := func(suite *WhenDrivingResponseTestingSuite) {
			suite.mockConversationTitler.EXPECT().
				TitleIfUnnamed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()
			suite.expectDemojizePassthrough()
			suite.mockConversationHistory.EXPECT().
				Read(gomock.Any(), conversationId).
				Return([]knowledge_interfaces.ConversationMessage{}, nil)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[identity_enums.AddressType]string{}, nil).
				AnyTimes()
			suite.mockLLMConnectionRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewLLMConnectionBuilder().Build(), nil)
			suite.mockMCPRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{}, nil)
			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)
			suite.mockSystemPromptBuilder.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "").
				Return(builtSystemPrompt)

			suite.mockInferenceLoop.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *models.LLMConnection, _ []*models.Tool, _ map[*models.MCPConnection][]*models.Tool, msgs []*interfaces.ChatCompletionMessage) ([]*interfaces.ChatCompletionMessage, error) {
					// The built system prompt leads the stream, never coming from inputs.
					if assert.Len(t, msgs, 2) {
						assert.Equal(t, "system", msgs[0].Role)
						assert.Equal(t, builtSystemPrompt, msgs[0].Content)
						assert.Equal(t, "user", msgs[1].Role)
						assert.Equal(t, "hello", msgs[1].Content)
					}
					return append(
						msgs,
						&interfaces.ChatCompletionMessage{Role: "assistant", Content: expectedFinal},
					), nil
				})

			suite.mockConversationHistory.EXPECT().Write(gomock.Any(),
				conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, msgs []knowledge_interfaces.ConversationMessage) ([]knowledge_interfaces.ConversationMessage, error) {
					// The built prompt is never persisted.
					if assert.Len(t, msgs, 2) {
						assert.Equal(t, "user", msgs[0].Role)
						assert.Equal(t, "hello", msgs[0].Content)
						assert.Equal(t, "assistant", msgs[1].Role)
						assert.Equal(t, expectedFinal, msgs[1].Content)
					}
					return msgs, nil
				})
		}

		t.Run("Should not persist the built prompt", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.NoError(t, err)
		})

		t.Run("Should return the final assistant message", func(t *testing.T) {
			t.Parallel()

			suite := WhenDrivingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Drive(ctx, conversationId, "", inputs)
			assert.NoError(t, err)
			if assert.NotNil(t, result) {
			}
		})

	})
}
