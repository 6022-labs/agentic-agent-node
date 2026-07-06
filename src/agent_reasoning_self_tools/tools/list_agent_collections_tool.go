package tools

import (
	"context"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"

	identity_requests "github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
)

// ListAgentCollectionsTool exposes the ListAgentCollections use case as a built-in self tool.
type ListAgentCollectionsTool struct {
	listAgentCollectionsExecutor tool_interfaces.ListAgentCollectionsExecutorInterface
}

func NewListAgentCollectionsTool(listAgentCollectionsExecutor tool_interfaces.ListAgentCollectionsExecutorInterface) *ListAgentCollectionsTool {
	return &ListAgentCollectionsTool{listAgentCollectionsExecutor: listAgentCollectionsExecutor}
}

func (t *ListAgentCollectionsTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "list_collections",
		Description: "List the governed agent collections on a chain.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"chainId": utils.StringProperty("Chain id to list collections for (e.g. 80002)."),
		}, []string{"chainId"}),
	}
}

func (t *ListAgentCollectionsTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	chainId, err := utils.StringArg(args, "chainId")
	if err != nil {
		return nil, err
	}
	return utils.ToToolResult(t.listAgentCollectionsExecutor.Execute(ctx, &identity_requests.ListAgentCollectionsRequest{ChainId: chainId}))
}
