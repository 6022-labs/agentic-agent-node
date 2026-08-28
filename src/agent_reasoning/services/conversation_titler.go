package services

import (
	"context"
	"fmt"
	"strings"

	knowledge_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

const titlePrompt = "You name conversations. Reply with a 3 to 6 word title for the conversation that starts " +
	"with the next message. Title only: no quotes, no trailing punctuation, no explanation."

// conversationTitleMaxLength matches the column the label is stored in.
const conversationTitleMaxLength = 128

// titleMaxOutputTokens bounds spend, not length: sanitizeTitle still enforces conversationTitleMaxLength.
const titleMaxOutputTokens = 32

// ConversationTitler asks the model for a label in one round-trip, with no tools:
// labelling must never act on the agent's behalf.
type ConversationTitler struct {
	chatCompletionRequester interfaces.ChatCompletionRequesterInterface
	conversationLabeler     knowledge_interfaces.ConversationLabelerInterface
	reasoningMetrics        metrics_interfaces.ReasoningMetricsInterface
}

func NewConversationTitler(
	chatCompletionRequester interfaces.ChatCompletionRequesterInterface,
	conversationLabeler knowledge_interfaces.ConversationLabelerInterface,
	reasoningMetrics metrics_interfaces.ReasoningMetricsInterface,
) *ConversationTitler {
	return &ConversationTitler{
		chatCompletionRequester: chatCompletionRequester,
		conversationLabeler:     conversationLabeler,
		reasoningMetrics:        reasoningMetrics,
	}
}

func (t *ConversationTitler) TitleIfUnnamed(
	ctx context.Context,
	conversationId string,
	llmConnection *models.LLMConnection,
	firstUserMessage string,
) error {
	firstUserMessage = strings.TrimSpace(firstUserMessage)
	if conversationId == "" || firstUserMessage == "" || llmConnection == nil {
		t.reasoningMetrics.RecordConversationTitled(ctx, metrics_interfaces.ConversationTitleOutcomeSkipped)
		return nil
	}

	name, err := t.conversationLabeler.GetName(ctx, conversationId)
	if err != nil {
		t.reasoningMetrics.RecordConversationTitled(ctx, metrics_interfaces.ConversationTitleOutcomeFailed)
		return fmt.Errorf("failed to read conversation name: %w", err)
	}
	if name != nil {
		t.reasoningMetrics.RecordConversationTitled(ctx, metrics_interfaces.ConversationTitleOutcomeSkipped)
		return nil
	}

	completion, err := t.chatCompletionRequester.ChatCompletion(
		ctx,
		llmConnection,
		nil,
		nil,
		[]*interfaces.ChatCompletionMessage{
			{Role: "system", Content: titlePrompt},
			{Role: "user", Content: firstUserMessage},
		},
		lo.ToPtr(titleMaxOutputTokens),
	)
	if err != nil {
		t.reasoningMetrics.RecordConversationTitled(ctx, metrics_interfaces.ConversationTitleOutcomeFailed)
		return fmt.Errorf("failed to generate conversation title: %w", err)
	}

	title := t.sanitizeTitle(completion)
	if title == "" {
		t.reasoningMetrics.RecordConversationTitled(ctx, metrics_interfaces.ConversationTitleOutcomeSkipped)
		return nil
	}

	if _, err := t.conversationLabeler.SetAvailableName(ctx, conversationId, title); err != nil {
		t.reasoningMetrics.RecordConversationTitled(ctx, metrics_interfaces.ConversationTitleOutcomeFailed)
		return fmt.Errorf("failed to store conversation title: %w", err)
	}

	t.reasoningMetrics.RecordConversationTitled(ctx, metrics_interfaces.ConversationTitleOutcomeTitled)
	return nil
}

// sanitizeTitle drops a uuid-shaped answer: it would collide with the session id callers scan replies for.
func (t *ConversationTitler) sanitizeTitle(completion *interfaces.LLMCompletion) string {
	if completion == nil || len(completion.Texts) == 0 {
		return ""
	}

	title := strings.Join(strings.Fields(completion.Texts[0]), " ")
	title = strings.Trim(title, `"'`)
	if runes := []rune(title); len(runes) > conversationTitleMaxLength {
		title = strings.TrimSpace(string(runes[:conversationTitleMaxLength]))
	}
	if _, err := uuid.Parse(title); err == nil {
		return ""
	}
	return title
}
