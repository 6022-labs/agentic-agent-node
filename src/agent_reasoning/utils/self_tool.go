package utils

import (
	"fmt"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

// SelfToolNamespace prefixes built-in tool function names; MCP connections may not claim it.
const SelfToolNamespace = "self"

func SelfToolToFunctionName(selfTool *models.Tool) string {
	return ToFunctionName(SelfToolNamespace, selfTool.Name)
}

// ExtractSelfToolNameFromFunctionName returns the bare tool name of a self-namespaced function name.
func ExtractSelfToolNameFromFunctionName(functionName string) (string, error) {
	toolName, found := strings.CutPrefix(functionName, SelfToolNamespace+FunctionNameSeparator)
	if !found {
		return "", fmt.Errorf("function name %q is outside the self namespace", functionName)
	}
	if toolName == "" {
		return "", fmt.Errorf("invalid function name format: %q", functionName)
	}
	return toolName, nil
}
