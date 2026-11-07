package services

import (
	"context"
	stderrors "errors"
	"strings"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_services_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	"github.com/samber/lo"
)

// MessageAppender demojizes content, stamps conversationId, hashes, drops
// dupes against stored rows, then inserts the rest. It materializes the
// agent-scoped Conversation head before the first append.
type MessageAppender struct {
	conversationMessageRepository repositories.ConversationMessageRepositoryInterface
	conversationRepository        repositories.ConversationRepositoryInterface
	agentRepository               identity_repositories.AgentRepositoryInterface
	knowledgeMetrics              metrics_interfaces.KnowledgeMetricsInterface
	emojiDecoder                  runtime_services_interfaces.EmojiDecoderInterface
}

func NewMessageAppender(
	conversationMessageRepository repositories.ConversationMessageRepositoryInterface,
	conversationRepository repositories.ConversationRepositoryInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
	knowledgeMetrics metrics_interfaces.KnowledgeMetricsInterface,
	emojiDecoder runtime_services_interfaces.EmojiDecoderInterface,
) *MessageAppender {
	return &MessageAppender{
		conversationMessageRepository: conversationMessageRepository,
		conversationRepository:        conversationRepository,
		agentRepository:               agentRepository,
		knowledgeMetrics:              knowledgeMetrics,
		emojiDecoder:                  emojiDecoder,
	}
}

func (s *MessageAppender) Append(
	ctx context.Context,
	conversationId string,
	messages []models.ConversationMessage,
) ([]models.ConversationMessage, error) {
	conversationId = strings.TrimSpace(conversationId)
	if len(conversationId) == 0 || len(messages) == 0 {
		return nil, nil
	}

	// The transcript never holds system rows.
	for _, msg := range messages {
		if msg.Role == enums.ConversationMessageRoleSystem {
			return nil, errors.NewInternalError(
				"append_failed",
				stderrors.New("system messages must never be persisted"),
			)
		}
	}

	agent, err := s.agentRepository.Get(ctx)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, errors.NewInternalError("append_failed", errors.ErrAgentNotInitialized)
	}
	created, err := s.conversationRepository.Ensure(ctx, &models.Conversation{
		Id:      conversationId,
		AgentId: agent.Id,
	})
	if err != nil {
		return nil, err
	}
	if created {
		s.knowledgeMetrics.RecordConversationStarted(ctx)
	}

	prepared := make([]models.ConversationMessage, 0, len(messages))
	hashes := make([]string, 0, len(messages))
	for _, msg := range messages {
		msg.ConversationId = conversationId
		msg.Content = lo.EmptyableToPtr(s.emojiDecoder.DemojizeContent(lo.FromPtr(msg.Content)))

		if err := msg.ComputeHash(); err != nil {
			return nil, err
		}

		prepared = append(prepared, msg)
		hashes = append(hashes, msg.Hash)
	}

	existing, err := s.conversationMessageRepository.ListExistingHashes(ctx, hashes)
	if err != nil {
		return nil, err
	}
	existingSet := make(map[string]struct{}, len(existing))
	for _, h := range existing {
		existingSet[h] = struct{}{}
	}

	toInsert := make([]models.ConversationMessage, 0, len(prepared))
	for _, msg := range prepared {
		if _, hit := existingSet[msg.Hash]; hit {
			continue
		}
		toInsert = append(toInsert, msg)
	}
	if len(toInsert) == 0 {
		return nil, nil
	}

	inserted, err := s.conversationMessageRepository.AppendMessages(ctx, toInsert)
	if err != nil {
		return nil, err
	}

	byRole := make(map[string]int64, len(inserted))
	for _, msg := range inserted {
		byRole[msg.Role.String()]++
	}
	for role, count := range byRole {
		s.knowledgeMetrics.RecordMessageAppended(ctx, role, count)
	}

	return inserted, nil
}
