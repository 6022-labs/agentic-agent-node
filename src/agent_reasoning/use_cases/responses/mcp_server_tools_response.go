package responses

import "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"

type MCPServerToolResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type MCPServerToolsResponse struct {
	Tools []MCPServerToolResponse `json:"tools"`
}

func ToMCPServerToolsResponse(tools []*models.Tool) *MCPServerToolsResponse {
	out := make([]MCPServerToolResponse, 0, len(tools))
	for _, tool := range tools {
		out = append(out, MCPServerToolResponse{Name: tool.Name, Description: tool.Description})
	}
	return &MCPServerToolsResponse{Tools: out}
}
