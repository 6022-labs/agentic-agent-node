package models

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
)

type ExternalAnthropicMessagesStandardRequest struct {
	Model     string `json:"model"`
	MaxTokens uint   `json:"max_tokens"`

	Tools      []ExternalAnthropicMessagesStandardRequestTool      `json:"tools,omitempty"`
	System     []ExternalAnthropicMessagesStandardRequestSystem    `json:"system,omitempty"`
	Messages   []ExternalAnthropicMessagesStandardMessage          `json:"messages,omitempty"`
	McpServers []ExternalAnthropicMessagesStandardRequestMCPServer `json:"mcp_servers,omitempty"`
}

type ExternalAnthropicMessagesStandardRequestTool struct {
	Name        string                                                  `json:"name"`
	Description string                                                  `json:"description"`
	InputSchema ExternalAnthropicMessagesStandardRequestToolInputSchema `json:"input_schema"`
}

type ExternalAnthropicMessagesStandardRequestToolInputSchema struct {
	Type       string                            `json:"type"`
	Required   []string                          `json:"required"`
	Properties ExternalToolInputSchemaProperties `json:"properties"`
}

type ExternalAnthropicMessagesStandardRequestSystem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ExternalAnthropicMessagesStandardRequestMCPServer struct {
	Type               string                                                              `json:"type"`
	Url                string                                                              `json:"url"`
	Name               string                                                              `json:"name"`
	AuthorizationToken *string                                                             `json:"authorization_token,omitempty"`
	ToolConfiguration  *ExternalAnthropicMessagesStandardRequestMCPServerToolConfiguration `json:"tool_configuration,omitempty"`
}

type ExternalAnthropicMessagesStandardRequestMCPServerToolConfiguration struct {
	AllowedTools []string `json:"allowed_tools,omitempty"`
}

func NewExternalAnthropicMessagesStandardRequest(
	model string,
	messages []ExternalAnthropicMessagesStandardMessage,
	system []ExternalAnthropicMessagesStandardRequestSystem,
) *ExternalAnthropicMessagesStandardRequest {
	return &ExternalAnthropicMessagesStandardRequest{
		Model:     model,
		MaxTokens: 32000,
		Messages:  messages,
		System:    system,
	}
}

// AddMcpConnectionAsMcpServer hands the whole server to the provider; a restricted connection carries allowed_tools.
func (r *ExternalAnthropicMessagesStandardRequest) AddMcpConnectionAsMcpServer(
	mcpConnection *models.MCPConnection,
	allowedToolNames []string,
) {
	authToken, _ := strings.CutPrefix(mcpConnection.HeaderAuth["Authorization"], "Bearer ")

	if r.McpServers == nil {
		r.McpServers = make([]ExternalAnthropicMessagesStandardRequestMCPServer, 0)
	}

	newMcpServer := ExternalAnthropicMessagesStandardRequestMCPServer{
		Type: "url",
		Url:  mcpConnection.Url,
		Name: mcpConnection.Name,
	}

	if len(strings.TrimSpace(authToken)) != 0 {
		newMcpServer.AuthorizationToken = &authToken
	}

	if len(mcpConnection.AllowedToolPatterns) > 0 {
		newMcpServer.ToolConfiguration = &ExternalAnthropicMessagesStandardRequestMCPServerToolConfiguration{
			AllowedTools: allowedToolNames,
		}
	}

	r.McpServers = append(r.McpServers, newMcpServer)
}

func (r *ExternalAnthropicMessagesStandardRequest) AddMcpConnectionAsFunctionCallTools(
	mcpConnection *models.MCPConnection,
	relatedTools []*models.Tool,
) {
	if len(relatedTools) == 0 {
		return
	}

	if r.Tools == nil {
		r.Tools = make([]ExternalAnthropicMessagesStandardRequestTool, 0)
	}

	for _, mcpTool := range relatedTools {
		r.appendFunctionCallTool(utils.McpToolToFunctionName(mcpConnection, mcpTool), mcpTool)
	}
}

func (r *ExternalAnthropicMessagesStandardRequest) AddSelfToolsAsFunctionCallTools(selfTools []*models.Tool) {
	if len(selfTools) == 0 {
		return
	}

	if r.Tools == nil {
		r.Tools = make([]ExternalAnthropicMessagesStandardRequestTool, 0)
	}

	for _, selfTool := range selfTools {
		r.appendFunctionCallTool(utils.SelfToolToFunctionName(selfTool), selfTool)
	}
}

func (r *ExternalAnthropicMessagesStandardRequest) appendFunctionCallTool(toolName string, tool *models.Tool) {
	required := tool.InputScheme.Required
	if required == nil {
		required = []string{}
	}

	r.Tools = append(r.Tools, ExternalAnthropicMessagesStandardRequestTool{
		Name:        toolName,
		Description: tool.Description,
		InputSchema: ExternalAnthropicMessagesStandardRequestToolInputSchema{
			Type:     tool.InputScheme.Type,
			Required: required,

			Properties: NewExternalToolInputSchemaPropertiesFromInternalModels(tool.InputScheme.Properties),
		},
	})
}
