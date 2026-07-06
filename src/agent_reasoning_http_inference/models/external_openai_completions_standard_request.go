package models

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
)

type ExternalOpenAICompletionsStandardRequest struct {
	Model    string                                            `json:"model"`
	Messages []ExternalOpenAICompletionsStandardRequestMessage `json:"messages"`
	Tools    []ExternalOpenAICompletionsStandardRequestTool    `json:"tools,omitempty"`
}

type ExternalOpenAICompletionsStandardRequestMessage struct {
	// Common fields
	Role    string  `json:"role"`
	Content *string `json:"content,omitempty"`

	// `Tool call` fields
	ToolCalls []ExternalOpenAICompletionsStandardRequestMessageToolCall `json:"tool_calls,omitempty"`

	// `Tool call output` fields
	ToolCallId *string `json:"tool_call_id,omitempty"`
}

type ExternalOpenAICompletionsStandardRequestMessageToolCall struct {
	ID       string                                            `json:"id"`
	Type     string                                            `json:"type"` // typically "function"
	Function ExternalOpenAICompletionsStandardToolCallFunction `json:"function"`
}

type ExternalOpenAICompletionsStandardRequestTool struct {
	Type     string                                               `json:"type"`
	Function ExternalOpenAICompletionsStandardRequestToolFunction `json:"function"`
}

type ExternalOpenAICompletionsStandardRequestToolFunction struct {
	Name        string                                                         `json:"name"`
	Strict      *bool                                                          `json:"strict,omitempty"`
	Description string                                                         `json:"description"`
	Parameters  ExternalOpenAICompletionsStandardRequestToolFunctionParameters `json:"parameters"`
}

type ExternalOpenAICompletionsStandardRequestToolFunctionParameters struct {
	Type                 string                            `json:"type"`
	Properties           ExternalToolInputSchemaProperties `json:"properties"`
	Required             []string                          `json:"required"`
	AdditionalProperties bool                              `json:"additionalProperties,omitempty"`
}

func NewExternalOpenAICompletionsStandardRequest(
	model string,
	messages []ExternalOpenAICompletionsStandardRequestMessage,
) *ExternalOpenAICompletionsStandardRequest {
	return &ExternalOpenAICompletionsStandardRequest{
		Model:    model,
		Messages: messages,
	}
}

func (r *ExternalOpenAICompletionsStandardRequest) AddMcpConnectionAsFunctionCallTools(
	mcpConnection *models.MCPConnection,
	relatedTools []*models.Tool,
) {
	if len(relatedTools) == 0 {
		return
	}

	if r.Tools == nil {
		r.Tools = make([]ExternalOpenAICompletionsStandardRequestTool, 0)
	}

	for _, mcpTool := range relatedTools {
		r.appendFunctionCallTool(utils.McpToolToFunctionName(mcpConnection, mcpTool), mcpTool)
	}
}

func (r *ExternalOpenAICompletionsStandardRequest) AddSelfToolsAsFunctionCallTools(selfTools []*models.Tool) {
	if len(selfTools) == 0 {
		return
	}

	if r.Tools == nil {
		r.Tools = make([]ExternalOpenAICompletionsStandardRequestTool, 0)
	}

	for _, selfTool := range selfTools {
		r.appendFunctionCallTool(utils.SelfToolToFunctionName(selfTool), selfTool)
	}
}

func (r *ExternalOpenAICompletionsStandardRequest) appendFunctionCallTool(name string, tool *models.Tool) {
	// MCPs are not always as well defined, so we set strict to false by default
	// But also not always all fields are required
	strict := false

	required := tool.InputScheme.Required
	if required == nil {
		required = []string{}
	}

	r.Tools = append(r.Tools, ExternalOpenAICompletionsStandardRequestTool{
		Type: "function",
		Function: ExternalOpenAICompletionsStandardRequestToolFunction{
			Name:        name,
			Strict:      &strict,
			Description: tool.Description,
			Parameters: ExternalOpenAICompletionsStandardRequestToolFunctionParameters{
				Type:                 tool.InputScheme.Type,
				Required:             required,
				AdditionalProperties: tool.InputScheme.AdditionalProperties,

				Properties: NewExternalToolInputSchemaPropertiesFromInternalModels(tool.InputScheme.Properties),
			},
		},
	})
}
