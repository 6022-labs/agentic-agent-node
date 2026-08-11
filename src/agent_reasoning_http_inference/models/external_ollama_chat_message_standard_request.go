package models

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
)

type ExternalOllamaChatMessageStandardRequest struct {
	Model    string                                           `json:"model"`
	Messages []ExternalOllamaChatMessageStandardMessage       `json:"messages"`
	Tools    []ExternalOllamaChatMessageStandardRequestTool   `json:"tools,omitempty"`
	Think    *bool                                            `json:"think,omitempty"`
	Stream   *bool                                            `json:"stream,omitempty"`
	Options  *ExternalOllamaChatMessageStandardRequestOptions `json:"options,omitempty"`
}

type ExternalOllamaChatMessageStandardRequestOptions struct {
	NumPredict *int `json:"num_predict,omitempty"`
}

type ExternalOllamaChatMessageStandardRequestTool struct {
	Type     string                                                   `json:"type"`
	Function ExternalOllamaChatMessageStandardMessageToolCallFunction `json:"function"`
}

func NewExternalOllamaChatMessageStandardRequest(
	model string,
	messages []ExternalOllamaChatMessageStandardMessage,
) *ExternalOllamaChatMessageStandardRequest {
	stream := false

	return &ExternalOllamaChatMessageStandardRequest{
		Model:    model,
		Stream:   &stream,
		Messages: messages,
	}
}

func (r *ExternalOllamaChatMessageStandardRequest) AddMcpConnectionAsFunctionCallTools(
	mcpConnection *models.MCPConnection,
	tools []*models.Tool,
) {
	if len(tools) == 0 {
		return
	}

	if r.Tools == nil {
		r.Tools = make([]ExternalOllamaChatMessageStandardRequestTool, 0)
	}

	for _, tool := range tools {
		r.appendFunctionCallTool(utils.McpToolToFunctionName(mcpConnection, tool), tool)
	}
}

func (r *ExternalOllamaChatMessageStandardRequest) AddSelfToolsAsFunctionCallTools(selfTools []*models.Tool) {
	if len(selfTools) == 0 {
		return
	}

	if r.Tools == nil {
		r.Tools = make([]ExternalOllamaChatMessageStandardRequestTool, 0)
	}

	for _, selfTool := range selfTools {
		r.appendFunctionCallTool(utils.SelfToolToFunctionName(selfTool), selfTool)
	}
}

func (r *ExternalOllamaChatMessageStandardRequest) appendFunctionCallTool(name string, tool *models.Tool) {
	required := tool.InputScheme.Required
	if required == nil {
		required = []string{}
	}

	r.Tools = append(r.Tools, ExternalOllamaChatMessageStandardRequestTool{
		Type: "function",
		Function: ExternalOllamaChatMessageStandardMessageToolCallFunction{
			Name:        name,
			Description: &tool.Description,
			Parameters: &ExternalOllamaChatMessageStandardMessageToolCallFunctionArguments{
				Type:       tool.InputScheme.Type,
				Required:   required,
				Properties: NewExternalToolInputSchemaPropertiesFromInternalModels(tool.InputScheme.Properties),
			},
		},
	})
}
