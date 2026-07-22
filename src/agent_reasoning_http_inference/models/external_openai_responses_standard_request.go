package models

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"github.com/samber/lo"
)

type ExternalOpenAIResponsesStandardRequest struct {
	Model              string                                        `json:"model"`
	Input              []ExternalOpenAIResponsesStandardRequestInput `json:"input"`
	Tools              []ExternalOpenAIResponsesStandardRequestTool  `json:"tools"`
	ToolChoice         string                                        `json:"tool_choice,omitempty"`
	PreviousResponseID *string                                       `json:"previous_response_id,omitempty"`
}

type ExternalOpenAIResponsesStandardRequestInput struct {
	Type string `json:"type"`

	// Type `message` fields
	Role    *string `json:"role,omitempty"`
	Content *string `json:"content,omitempty"`

	// Type `function_call` fields
	Name      *string `json:"name,omitempty"`
	Arguments *string `json:"arguments,omitempty"`

	// Type `function_call` and `function_call_output` fields
	CallId *string `json:"call_id,omitempty"`

	// Type `function_call_output` fields
	Output *string `json:"output,omitempty"`
}

type ExternalOpenAIResponsesStandardToolRequireApproval string

const (
	ExternalOpenAIResponsesStandardToolRequireApprovalNever  ExternalOpenAIResponsesStandardToolRequireApproval = "never"
	ExternalOpenAIResponsesStandardToolRequireApprovalAlways ExternalOpenAIResponsesStandardToolRequireApproval = "always"
)

type ExternalOpenAIResponsesStandardRequestTool struct {
	// Common fields
	Type string `json:"type"`

	// Function type tool fields
	Name        *string                                                       `json:"name,omitempty"`
	Strict      *bool                                                         `json:"strict,omitempty"`
	Parameters  *ExternalOpenAIResponsesStandardRequestToolFunctionParameters `json:"parameters,omitempty"`
	Description *string                                                       `json:"description,omitempty"`

	// Mcp type tool fields
	Headers         *map[string]string                                  `json:"headers,omitempty"`
	ServerUrl       *string                                             `json:"server_url,omitempty"`
	ServerLabel     *string                                             `json:"server_label,omitempty"`
	RequireApproval *ExternalOpenAIResponsesStandardToolRequireApproval `json:"require_approval,omitempty"`
}

type ExternalOpenAIResponsesStandardRequestToolFunctionParameters struct {
	Type                 string                            `json:"type"`
	Properties           ExternalToolInputSchemaProperties `json:"properties"`
	Required             []string                          `json:"required"`
	AdditionalProperties bool                              `json:"additionalProperties"`
}

func NewExternalOpenAIResponsesStandardRequest(
	model string,
	inputs []ExternalOpenAIResponsesStandardRequestInput,
	previousResponseId *string,
) *ExternalOpenAIResponsesStandardRequest {
	return &ExternalOpenAIResponsesStandardRequest{
		Model:              model,
		ToolChoice:         "auto",
		Input:              inputs,
		PreviousResponseID: previousResponseId,
	}
}

func (r *ExternalOpenAIResponsesStandardRequest) AddMcpConnectionsAsMcpTools(
	mcpConnections []*models.MCPConnection,
) {
	if len(mcpConnections) == 0 {
		return
	}

	if r.Tools == nil {
		r.Tools = make([]ExternalOpenAIResponsesStandardRequestTool, 0)
	}

	for _, conn := range mcpConnections {
		r.Tools = append(r.Tools, ExternalOpenAIResponsesStandardRequestTool{
			Type: "mcp",

			ServerLabel: &conn.Name,
			Headers:     &conn.HeaderAuth,

			ServerUrl:       &conn.Url,
			RequireApproval: lo.ToPtr(ExternalOpenAIResponsesStandardToolRequireApprovalNever),
		})
	}
}

func (r *ExternalOpenAIResponsesStandardRequest) AddMcpConnectionAsFunctionCallTools(
	mcpConnection *models.MCPConnection,
	relatedTools []*models.Tool,
) {
	if len(relatedTools) == 0 {
		return
	}

	if r.Tools == nil {
		r.Tools = make([]ExternalOpenAIResponsesStandardRequestTool, 0)
	}

	for _, mcpTool := range relatedTools {
		r.appendFunctionCallTool(utils.McpToolToFunctionName(mcpConnection, mcpTool), mcpTool)
	}
}

func (r *ExternalOpenAIResponsesStandardRequest) AddSelfToolsAsFunctionCallTools(selfTools []*models.Tool) {
	if len(selfTools) == 0 {
		return
	}

	if r.Tools == nil {
		r.Tools = make([]ExternalOpenAIResponsesStandardRequestTool, 0)
	}

	for _, selfTool := range selfTools {
		r.appendFunctionCallTool(utils.SelfToolToFunctionName(selfTool), selfTool)
	}
}

func (r *ExternalOpenAIResponsesStandardRequest) appendFunctionCallTool(name string, tool *models.Tool) {
	// MCPs are not always as well defined, so we set strict to false by default
	// But also not always all fields are required
	strict := false

	toolToAdd := ExternalOpenAIResponsesStandardRequestTool{
		Type:        "function",
		Name:        &name,
		Strict:      &strict,
		Description: &tool.Description,
		Parameters: &ExternalOpenAIResponsesStandardRequestToolFunctionParameters{
			Type:                 tool.InputScheme.Type,
			Required:             tool.InputScheme.Required,
			AdditionalProperties: tool.InputScheme.AdditionalProperties,

			Properties: NewExternalToolInputSchemaPropertiesFromInternalModels(tool.InputScheme.Properties),
		},
	}

	if toolToAdd.Parameters.Required == nil {
		toolToAdd.Parameters.Required = make([]string, 0)
	}

	r.Tools = append(r.Tools, toolToAdd)
}
