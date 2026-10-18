package tools

import (
	"context"

	knowledge_requests "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/samber/lo"
)

// ListConversationsTool lists the caller's sessions (all of them for the agent owner).
type ListConversationsTool struct {
	listConversationsExecutor tool_interfaces.ListConversationsExecutorInterface
}

func NewListConversationsTool(
	listConversationsExecutor tool_interfaces.ListConversationsExecutorInterface,
) *ListConversationsTool {
	return &ListConversationsTool{listConversationsExecutor: listConversationsExecutor}
}

func (t *ListConversationsTool) Definition() *models.Tool {
	return &models.Tool{
		Name: "list_conversations",
		Description: "List the caller's memory sessions (all of them for the agent owner): id, name, message " +
			"count, last activity, most recent first. Quote a session's conversationId verbatim whenever you are asked " +
			"which session a conversation belongs to.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"page":            utils.IntegerProperty("Page number, 1-based. Defaults to 1."),
			"elementsPerPage": utils.IntegerProperty("Sessions per page, at most 100. Defaults to 10."),
		}, nil),
	}
}

func (t *ListConversationsTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	request := &knowledge_requests.ListConversationsRequest{}
	if page := utils.OptionalIntArg(args, "page"); page != nil && *page > 0 {
		request.Page = lo.ToPtr(uint(*page))
	}
	if elementsPerPage := utils.OptionalIntArg(
		args,
		"elementsPerPage",
	); elementsPerPage != nil &&
		*elementsPerPage > 0 {
		request.ElementsPerPage = lo.ToPtr(uint(*elementsPerPage))
	}
	return utils.ToToolResult(t.listConversationsExecutor.Execute(ctx, request))
}

func (t *ListConversationsTool) IsEnabledByDefault() bool {
	return true
}

func (t *ListConversationsTool) GroupName() string {
	return GroupMemory
}
