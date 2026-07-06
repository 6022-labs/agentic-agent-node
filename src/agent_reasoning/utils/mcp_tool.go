package utils

import (
	"fmt"
	"sort"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/samber/lo"
)

// McpToolNamespace prefixes MCP tool function names so they can never collide with self tools.
const McpToolNamespace = "mcp"

func McpToolToFunctionName(mcpConnection *models.MCPConnection, mcpTool *models.Tool) string {
	return ToFunctionName(McpToolNamespace, ToFunctionName(mcpConnection.Name, mcpTool.Name))
}

// SortedMcpConnections returns the map's connections in stable name order for deterministic rendering.
func SortedMcpConnections(mcpTools map[*models.MCPConnection][]*models.Tool) []*models.MCPConnection {
	mcpConnections := lo.Keys(mcpTools)
	sort.Slice(mcpConnections, func(i, j int) bool { return mcpConnections[i].Name < mcpConnections[j].Name })
	return mcpConnections
}

// FindMcpConnectionByName returns the connection with the given name, if present.
func FindMcpConnectionByName(mcpTools map[*models.MCPConnection][]*models.Tool, name string) (*models.MCPConnection, bool) {
	return lo.FindKeyBy(mcpTools, func(mcpConnection *models.MCPConnection, _ []*models.Tool) bool {
		return mcpConnection.Name == name
	})
}

func ExtractMcpInfoFromFunctionName(
	functionName string,
	mcpTools map[*models.MCPConnection][]*models.Tool,
) (*models.MCPConnection, *models.Tool, error) {
	namespaced, found := strings.CutPrefix(functionName, McpToolNamespace+FunctionNameSeparator)
	if !found {
		return nil, nil, fmt.Errorf("function name %q is outside the mcp namespace", functionName)
	}
	parts := strings.SplitN(namespaced, FunctionNameSeparator, 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return nil, nil, fmt.Errorf("invalid function name format: %q", functionName)
	}
	connName := parts[0]
	toolName := parts[1]

	mcpConnection, found := FindMcpConnectionByName(mcpTools, connName)
	if !found {
		return nil, nil, fmt.Errorf("no MCP connection found for function name: %q", functionName)
	}

	mcpTool, found := lo.Find(mcpTools[mcpConnection], func(mcpTool *models.Tool) bool {
		return mcpTool.Name == toolName
	})
	if !found {
		return nil, nil, fmt.Errorf("no MCP tool found for function name: %q", functionName)
	}

	return mcpConnection, mcpTool, nil
}
