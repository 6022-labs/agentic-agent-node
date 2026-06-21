package services

import (
	"context"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/samber/lo"
)

// ConversationHistory implements ConversationHistoryInterface, mapping
// between the port DTO and the persistence model.
type ConversationHistory struct {
	conversationMessageRepository repositories.ConversationMessageRepositoryInterface
	messageAppender               interfaces.MessageAppenderInterface
}

func NewConversationHistory(
	conversationMessageRepository repositories.ConversationMessageRepositoryInterface,
	messageAppender interfaces.MessageAppenderInterface,
) *ConversationHistory {
	return &ConversationHistory{
		conversationMessageRepository: conversationMessageRepository,
		messageAppender:               messageAppender,
	}
}

func (h *ConversationHistory) Read(ctx context.Context, conversationId string) ([]interfaces.ConversationMessage, error) {
	conversationId = strings.TrimSpace(conversationId)
	if len(conversationId) == 0 {
		return nil, nil
	}
	stored, err := h.conversationMessageRepository.ListByConversationId(ctx, conversationId)
	if err != nil {
		return nil, err
	}
	out := make([]interfaces.ConversationMessage, 0, len(stored))
	for _, m := range stored {
		out = append(out, h.toDTO(m))
	}
	return out, nil
}

func (h *ConversationHistory) Write(ctx context.Context, conversationId string, messages []interfaces.ConversationMessage) ([]interfaces.ConversationMessage, error) {
	in := make([]models.ConversationMessage, 0, len(messages))
	for _, m := range messages {
		in = append(in, h.toModel(m))
	}
	persisted, err := h.messageAppender.Append(ctx, conversationId, in)
	if err != nil {
		return nil, err
	}
	out := make([]interfaces.ConversationMessage, 0, len(persisted))
	for _, m := range persisted {
		out = append(out, h.toDTO(m))
	}
	return out, nil
}

func (h *ConversationHistory) toDTO(m models.ConversationMessage) interfaces.ConversationMessage {
	return interfaces.ConversationMessage{
		Role:       m.Role.String(),
		Content:    lo.FromPtr(m.Content),
		ToolCalls:  lo.FromPtr(m.ToolCalls),
		ToolCallId: lo.FromPtr(m.ToolCallId),
		Timestamp:  m.Timestamp,
	}
}

func (h *ConversationHistory) toModel(m interfaces.ConversationMessage) models.ConversationMessage {
	return models.ConversationMessage{
		Role:       enums.ConversationMessageRole(m.Role),
		Content:    lo.EmptyableToPtr(m.Content),
		ToolCalls:  lo.EmptyableToPtr(m.ToolCalls),
		ToolCallId: lo.EmptyableToPtr(m.ToolCallId),
		Timestamp:  m.Timestamp,
	}
}
