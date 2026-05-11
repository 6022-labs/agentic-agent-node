package services

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
)

// ConversationHistory implements ConversationHistoryInterface, mapping
// between the port DTO and the persistence model.
type ConversationHistory struct {
	listService   interfaces.ListConversationMessagesInterface
	appendService interfaces.AppendConversationMessagesInterface
}

func NewConversationHistory(
	listService interfaces.ListConversationMessagesInterface,
	appendService interfaces.AppendConversationMessagesInterface,
) *ConversationHistory {
	return &ConversationHistory{
		listService:   listService,
		appendService: appendService,
	}
}

func (h *ConversationHistory) Read(conversationId string) ([]interfaces.ConversationMessage, error) {
	stored, err := h.listService.List(conversationId)
	if err != nil {
		return nil, err
	}
	out := make([]interfaces.ConversationMessage, 0, len(stored))
	for _, m := range stored {
		out = append(out, h.toDTO(m))
	}
	return out, nil
}

func (h *ConversationHistory) Write(conversationId string, messages []interfaces.ConversationMessage) ([]interfaces.ConversationMessage, error) {
	in := make([]models.ConversationMessage, 0, len(messages))
	for _, m := range messages {
		in = append(in, h.toModel(m))
	}
	persisted, err := h.appendService.Append(conversationId, in)
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
		Role:      m.Role,
		Content:   m.Content,
		Timestamp: m.Timestamp,
	}
}

func (h *ConversationHistory) toModel(m interfaces.ConversationMessage) models.ConversationMessage {
	return models.ConversationMessage{
		Role:      m.Role,
		Content:   m.Content,
		Timestamp: m.Timestamp,
	}
}
