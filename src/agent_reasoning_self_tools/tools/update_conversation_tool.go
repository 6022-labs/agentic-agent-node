package tools

import (
	"context"

	knowledge_requests "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// UpdateConversationTool exposes the UpdateConversation use case as a built-in self tool.
type UpdateConversationTool struct {
	updateConversationExecutor tool_interfaces.UpdateConversationExecutorInterface
}

func NewUpdateConversationTool(
	updateConversationExecutor tool_interfaces.UpdateConversationExecutorInterface,
) *UpdateConversationTool {
	return &UpdateConversationTool{updateConversationExecutor: updateConversationExecutor}
}

func (t *UpdateConversationTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "update_conversation",
		Description: "Rename one of your memory sessions.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"conversationId": utils.StringProperty("Id of the session to rename."),
			"name":           utils.StringProperty("Short title for the session, 3 to 6 words. Must not be a uuid."),
		}, []string{"conversationId", "name"}),
	}
}

func (t *UpdateConversationTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	conversationId, err := utils.StringArg(args, "conversationId")
	if err != nil {
		return nil, err
	}
	name, err := utils.StringArg(args, "name")
	if err != nil {
		return nil, err
	}
	return utils.ToToolResult(
		t.updateConversationExecutor.Execute(ctx, &knowledge_requests.UpdateConversationRequest{
			ConversationId: conversationId,
			Name:           name,
		}),
	)
}

func (t *UpdateConversationTool) IsEnabledByDefault() bool {
	return true
}

func (t *UpdateConversationTool) GroupName() string {
	return GroupMemory
}
