package use_cases

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
)

type ListConversationMessagesInterface interface {
	Execute(conversationId string) *usecase_result.Result
}

type ListConversationMessages struct {
	service interfaces.ListConversationMessagesInterface
}

func NewListConversationMessages(service interfaces.ListConversationMessagesInterface) *ListConversationMessages {
	return &ListConversationMessages{service: service}
}

func (u *ListConversationMessages) Execute(conversationId string) *usecase_result.Result {
	messages, err := u.service.List(conversationId)
	if err != nil {
		return usecase_result.InternalError("list_failed", err.Error())
	}
	return usecase_result.Ok(&responses.ListConversationMessagesResponse{Messages: toWireMessages(messages)})
}
