package utils

import (
	"sort"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

// FunctionNameSeparator joins the namespace and tool segments of a function name.
const FunctionNameSeparator = "_"

func ToFunctionName(namespace, toolName string) string {
	return namespace + FunctionNameSeparator + toolName
}

// ExtractNamespaceFromFunctionName returns the namespace segment of a function name, or "" when there is none.
func ExtractNamespaceFromFunctionName(functionName string) string {
	namespace, _, found := strings.Cut(functionName, FunctionNameSeparator)
	if !found {
		return ""
	}
	return namespace
}

// SortedToolsByName returns a copy of the tools ordered by name for deterministic rendering and serialization.
func SortedToolsByName(tools []*models.Tool) []*models.Tool {
	sorted := make([]*models.Tool, len(tools))
	copy(sorted, tools)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	return sorted
}
