package tools

import (
	"context"

	knowledge_requests "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// CreateConversationTool exposes CreateConversation as a self tool, scoped to the turn's caller.
type CreateConversationTool struct {
	createConversationExecutor tool_interfaces.CreateConversationExecutorInterface
}

func NewCreateConversationTool(
	createConversationExecutor tool_interfaces.CreateConversationExecutorInterface,
) *CreateConversationTool {
	return &CreateConversationTool{createConversationExecutor: createConversationExecutor}
}

func (t *CreateConversationTool) Definition() *models.Tool {
	return &models.Tool{
		Name: "create_conversation",
		Description: "Open a new named memory session and return its conversationId. " +
			"Use it when no existing session fits the topic at hand.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"name": utils.StringProperty("Short title for the session, 3 to 6 words. Must not be a uuid."),
		}, []string{"name"}),
	}
}

func (t *CreateConversationTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	name, err := utils.StringArg(args, "name")
	if err != nil {
		return nil, err
	}
	return utils.ToToolResult(
		t.createConversationExecutor.Execute(ctx, &knowledge_requests.CreateConversationRequest{
			ConversationNameRequest: knowledge_requests.ConversationNameRequest{Name: name},
		}),
	)
}

func (t *CreateConversationTool) IsEnabledByDefault() bool {
	return true
}

func (t *CreateConversationTool) GroupName() string {
	return GroupMemory
}
