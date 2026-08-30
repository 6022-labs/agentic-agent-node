package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	knowledge_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/services_mocks/interfaces_mocks"

	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/metrics_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenTitlingConversationTestingSuite struct {
	sut *services.ConversationTitler

	mockChatCompletionRequester *interfaces_mocks.MockChatCompletionRequesterInterface
	mockConversationLabeler     *knowledge_interfaces_mocks.MockConversationLabelerInterface
	mockReasoningMetrics        *metrics_mocks.MockReasoningMetricsInterface
}

func WhenTitlingConversationBeforeEach(t *testing.T) *WhenTitlingConversationTestingSuite {
	mockController := gomock.NewController(t)

	mockChatCompletionRequester := interfaces_mocks.NewMockChatCompletionRequesterInterface(mockController)
	mockConversationLabeler := knowledge_interfaces_mocks.NewMockConversationLabelerInterface(mockController)
	mockReasoningMetrics := metrics_mocks.NewMockReasoningMetricsInterface(mockController)

	return &WhenTitlingConversationTestingSuite{
		sut: services.NewConversationTitler(
			mockChatCompletionRequester,
			mockConversationLabeler,
			mockReasoningMetrics,
		),
		mockChatCompletionRequester: mockChatCompletionRequester,
		mockConversationLabeler:     mockConversationLabeler,
		mockReasoningMetrics:        mockReasoningMetrics,
	}
}

func llmConnectionForTitling() *models.LLMConnection {
	return &models.LLMConnection{Model: "claude-opus-5"}
}

func TestWhenTitlingConversation(t *testing.T) {
	t.Parallel()

	t.Run("Given there is no opening message", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenTitlingConversationTestingSuite) {
			suite.mockReasoningMetrics.EXPECT().
				RecordConversationTitled(gomock.Any(), metrics_interfaces.ConversationTitleOutcomeSkipped)
		}

		t.Run("Should skip without touching the labeler", func(t *testing.T) {
			t.Parallel()

			suite := WhenTitlingConversationBeforeEach(t)
			initSuite(suite)

			err := suite.sut.TitleIfUnnamed(context.Background(), "c1", llmConnectionForTitling(), "   ")

			assert.NoError(t, err)
		})
	})

	t.Run("Given the conversation is already named", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenTitlingConversationTestingSuite) {
			suite.mockConversationLabeler.EXPECT().GetName(gomock.Any(), "c1").Return(lo.ToPtr("Already named"), nil)
			suite.mockReasoningMetrics.EXPECT().
				RecordConversationTitled(gomock.Any(), metrics_interfaces.ConversationTitleOutcomeSkipped)
		}

		t.Run("Should never call the model", func(t *testing.T) {
			t.Parallel()

			suite := WhenTitlingConversationBeforeEach(t)
			initSuite(suite)

			err := suite.sut.TitleIfUnnamed(context.Background(), "c1", llmConnectionForTitling(), "hello")

			assert.NoError(t, err)
		})
	})

	t.Run("Given the conversation has no name", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenTitlingConversationTestingSuite) {
			suite.mockConversationLabeler.EXPECT().GetName(gomock.Any(), "c1").Return(nil, nil)
		}

		t.Run("Should ask the model once, with no tools, and store the title", func(t *testing.T) {
			t.Parallel()

			suite := WhenTitlingConversationBeforeEach(t)
			initSuite(suite)

			suite.mockChatCompletionRequester.EXPECT().
				ChatCompletion(gomock.Any(), gomock.Any(), gomock.Nil(), gomock.Nil(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					_ *models.LLMConnection,
					_ []*models.Tool,
					_ map[*models.MCPConnection][]*models.Tool,
					messages []*interfaces.ChatCompletionMessage,
					maxOutputTokens *int,
				) (*interfaces.LLMCompletion, error) {
					assert.Len(t, messages, 2)
					assert.Equal(t, "system", messages[0].Role)
					assert.Equal(t, "review this ERC20 mint", messages[1].Content)
					// A handful of words never needs the provider ceiling.
					assert.Equal(t, 32, *maxOutputTokens)
					return &interfaces.LLMCompletion{Texts: []string{" \"ERC20 mint audit\"  "}}, nil
				})
			suite.mockConversationLabeler.EXPECT().
				SetAvailableName(gomock.Any(), "c1", "ERC20 mint audit").
				Return(nil, nil)
			suite.mockReasoningMetrics.EXPECT().
				RecordConversationTitled(gomock.Any(), metrics_interfaces.ConversationTitleOutcomeTitled)

			err := suite.sut.TitleIfUnnamed(
				context.Background(),
				"c1",
				llmConnectionForTitling(),
				"review this ERC20 mint",
			)

			assert.NoError(t, err)
		})

		t.Run("Should truncate a title the column could not hold", func(t *testing.T) {
			t.Parallel()

			suite := WhenTitlingConversationBeforeEach(t)
			initSuite(suite)

			// The token cap bounds spend, not characters, so an over-long answer is still possible.
			suite.mockChatCompletionRequester.EXPECT().
				ChatCompletion(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{Texts: []string{strings.Repeat("a", 200)}}, nil)
			suite.mockConversationLabeler.EXPECT().
				SetAvailableName(gomock.Any(), "c1", gomock.Any()).
				Do(func(_ context.Context, _ string, name string) {
					assert.Len(t, []rune(name), 128)
				}).
				Return(nil, nil)
			suite.mockReasoningMetrics.EXPECT().
				RecordConversationTitled(gomock.Any(), metrics_interfaces.ConversationTitleOutcomeTitled)

			err := suite.sut.TitleIfUnnamed(context.Background(), "c1", llmConnectionForTitling(), "hello")

			assert.NoError(t, err)
		})

		t.Run("Should skip a uuid-shaped title", func(t *testing.T) {
			t.Parallel()

			suite := WhenTitlingConversationBeforeEach(t)
			initSuite(suite)

			suite.mockChatCompletionRequester.EXPECT().
				ChatCompletion(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{Texts: []string{uuid.NewString()}}, nil)
			suite.mockReasoningMetrics.EXPECT().
				RecordConversationTitled(gomock.Any(), metrics_interfaces.ConversationTitleOutcomeSkipped)

			err := suite.sut.TitleIfUnnamed(context.Background(), "c1", llmConnectionForTitling(), "hello")

			assert.NoError(t, err)
		})

		t.Run("Should report a failing model call without panicking", func(t *testing.T) {
			t.Parallel()

			suite := WhenTitlingConversationBeforeEach(t)
			initSuite(suite)

			suite.mockChatCompletionRequester.EXPECT().
				ChatCompletion(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("provider down"))
			suite.mockReasoningMetrics.EXPECT().
				RecordConversationTitled(gomock.Any(), metrics_interfaces.ConversationTitleOutcomeFailed)

			err := suite.sut.TitleIfUnnamed(context.Background(), "c1", llmConnectionForTitling(), "hello")

			assert.ErrorContains(t, err, "provider down")
		})
	})
}
