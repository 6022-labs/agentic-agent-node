package services

import (
	"context"
	"fmt"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/metrics/interfaces"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

// SelfToolRegistry aggregates the built-in self tools contributed to the dig group.
type SelfToolRegistry struct {
	logger          *zap.Logger
	selfToolMetrics metrics_interfaces.SelfToolMetricsInterface
	toolsByName     map[string]interfaces.SelfToolInterface
	definitions     []*models.Tool
	defaultEnabled  map[string]bool
	groupByName     map[string]string
}

type NewSelfToolRegistryParams struct {
	dig.In

	Logger          *zap.Logger
	SelfToolMetrics metrics_interfaces.SelfToolMetricsInterface
	SelfTools       []interfaces.SelfToolInterface `group:"self_tools"`
}

func NewSelfToolRegistry(params NewSelfToolRegistryParams) *SelfToolRegistry {
	toolsByName := make(map[string]interfaces.SelfToolInterface, len(params.SelfTools))
	definitions := make([]*models.Tool, 0, len(params.SelfTools))
	defaultEnabled := make(map[string]bool, len(params.SelfTools))
	groupByName := make(map[string]string, len(params.SelfTools))

	for _, tool := range params.SelfTools {
		definition := tool.Definition()
		if _, exists := toolsByName[definition.Name]; exists {
			params.Logger.Warn("duplicate self tool registration ignored", zap.String("toolName", definition.Name))
			continue
		}
		toolsByName[definition.Name] = tool
		definitions = append(definitions, definition)
		defaultEnabled[definition.Name] = tool.IsEnabledByDefault()
		groupByName[definition.Name] = tool.GroupName()
	}

	return &SelfToolRegistry{
		logger:          params.Logger,
		selfToolMetrics: params.SelfToolMetrics,
		toolsByName:     toolsByName,
		definitions:     utils.SortedToolsByName(definitions),
		defaultEnabled:  defaultEnabled,
		groupByName:     groupByName,
	}
}

func (r *SelfToolRegistry) ListTools() []*models.Tool {
	return r.definitions
}

func (r *SelfToolRegistry) ListDefaultEnabled() map[string]bool {
	return r.defaultEnabled
}

func (r *SelfToolRegistry) ListGroupNames() map[string]string {
	return r.groupByName
}

func (r *SelfToolRegistry) CallTool(ctx context.Context, toolName string, args map[string]any) (map[string]any, error) {
	tool, found := r.toolsByName[toolName]
	if !found {
		return nil, fmt.Errorf("unknown self tool: %q", toolName)
	}

	start := time.Now()
	result, err := tool.Execute(ctx, args)
	r.selfToolMetrics.RecordToolCall(ctx, toolName, time.Since(start), err)

	return result, err
}
