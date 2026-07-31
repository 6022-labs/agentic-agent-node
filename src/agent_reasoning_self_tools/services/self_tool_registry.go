package services

import (
	"context"
	"fmt"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

// SelfToolRegistry aggregates the built-in self tools contributed to the dig group.
type SelfToolRegistry struct {
	logger      *zap.Logger
	toolsByName map[string]interfaces.SelfToolInterface
	definitions []*models.Tool
}

type NewSelfToolRegistryParams struct {
	dig.In

	Logger    *zap.Logger
	SelfTools []interfaces.SelfToolInterface `group:"self_tools"`
}

func NewSelfToolRegistry(params NewSelfToolRegistryParams) *SelfToolRegistry {
	toolsByName := make(map[string]interfaces.SelfToolInterface, len(params.SelfTools))
	definitions := make([]*models.Tool, 0, len(params.SelfTools))

	for _, tool := range params.SelfTools {
		definition := tool.Definition()
		if _, exists := toolsByName[definition.Name]; exists {
			params.Logger.Warn("duplicate self tool registration ignored", zap.String("toolName", definition.Name))
			continue
		}
		toolsByName[definition.Name] = tool
		definitions = append(definitions, definition)
	}

	return &SelfToolRegistry{
		logger:      params.Logger,
		toolsByName: toolsByName,
		definitions: utils.SortedToolsByName(definitions),
	}
}

func (r *SelfToolRegistry) ListTools() []*models.Tool {
	return r.definitions
}

func (r *SelfToolRegistry) CallTool(ctx context.Context, toolName string, args map[string]any) (map[string]any, error) {
	tool, found := r.toolsByName[toolName]
	if !found {
		return nil, fmt.Errorf("unknown self tool: %q", toolName)
	}

	return tool.Execute(ctx, args)
}
