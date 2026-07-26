package responses

import "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"

type McpServerToolResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type McpServerToolsResponse struct {
	Tools []McpServerToolResponse `json:"tools"`
}

func ToMcpServerToolsResponse(p *models.ExternalMCPServerToolsPayload) *McpServerToolsResponse {
	out := make([]McpServerToolResponse, 0, len(p.Tools))
	for _, tool := range p.Tools {
		out = append(out, McpServerToolResponse{Name: tool.Name, Description: tool.Description})
	}
	return &McpServerToolsResponse{Tools: out}
}
