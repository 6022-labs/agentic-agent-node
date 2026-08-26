package services

import (
	"context"
	"fmt"
	"strings"

	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
)

// conversationNameMaxAttempts bounds the "name", "name 2", "name 3" … walk when a label is already taken.
const conversationNameMaxAttempts = 10

// ConversationLabeler is the single write path for a conversation's name; callers pick whether a
// collision is an error or something to work around.
type ConversationLabeler struct {
	conversationRepository repositories.ConversationRepositoryInterface
	knowledgeMetrics       metrics_interfaces.KnowledgeMetricsInterface
}

func NewConversationLabeler(
	conversationRepository repositories.ConversationRepositoryInterface,
	knowledgeMetrics metrics_interfaces.KnowledgeMetricsInterface,
) *ConversationLabeler {
	return &ConversationLabeler{
		conversationRepository: conversationRepository,
		knowledgeMetrics:       knowledgeMetrics,
	}
}

func (l *ConversationLabeler) GetName(ctx context.Context, conversationId string) (*string, error) {
	conversation, err := l.conversationRepository.GetById(ctx, strings.TrimSpace(conversationId))
	if err != nil {
		return nil, err
	}
	if conversation == nil {
		return nil, nil
	}
	return conversation.Name, nil
}

func (l *ConversationLabeler) SetName(
	ctx context.Context,
	conversationId string,
	name string,
) (*models.Conversation, error) {
	return l.apply(ctx, conversationId, name, false)
}

func (l *ConversationLabeler) SetAvailableName(
	ctx context.Context,
	conversationId string,
	name string,
) (*models.Conversation, error) {
	return l.apply(ctx, conversationId, name, true)
}

func (l *ConversationLabeler) apply(
	ctx context.Context,
	conversationId string,
	name string,
	suffixOnCollision bool,
) (*models.Conversation, error) {
	conversationId = strings.TrimSpace(conversationId)
	conversation, err := l.conversationRepository.GetById(ctx, conversationId)
	if err != nil {
		return nil, err
	}
	if conversation == nil {
		return nil, fmt.Errorf("conversation %q not found", conversationId)
	}

	available, err := l.resolveName(ctx, conversation, name, suffixOnCollision)
	if err != nil {
		return nil, err
	}

	conversation.Name = available
	updated, err := l.conversationRepository.Update(ctx, conversation)
	if err != nil {
		return nil, err
	}

	l.knowledgeMetrics.RecordConversationNamed(ctx)
	return updated, nil
}

// resolveName excludes the conversation itself, so renaming a session to the name it already holds is a no-op.
func (l *ConversationLabeler) resolveName(
	ctx context.Context,
	conversation *models.Conversation,
	name string,
	suffixOnCollision bool,
) (*string, error) {
	attempts := 1
	if suffixOnCollision {
		attempts = conversationNameMaxAttempts
	}

	for attempt := 1; attempt <= attempts; attempt++ {
		candidate := name
		if attempt > 1 {
			candidate = fmt.Sprintf("%s %d", name, attempt)
		}

		taken, err := l.conversationRepository.ExistsByAgentIdAndName(
			ctx,
			conversation.AgentId,
			candidate,
			&conversation.Id,
		)
		if err != nil {
			return nil, err
		}
		if !taken {
			return &candidate, nil
		}
	}

	return nil, interfaces.ErrConversationNameTaken
}
