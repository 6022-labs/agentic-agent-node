package models

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
)

type ExternalAmazonBedrockConverseStandardRequest struct {
	ModelId         string                                                       `json:"modelId"`
	Messages        []ExternalAmazonBedrockConverseStandardMessage               `json:"messages"`
	System          []ExternalAmazonBedrockConverseStandardRequestSystem         `json:"system"`
	ToolConfig      *ExternalAmazonBedrockConverseStandardRequestToolConfig      `json:"toolConfig,omitempty"`
	InferenceConfig *ExternalAmazonBedrockConverseStandardRequestInferenceConfig `json:"inferenceConfig,omitempty"`
}

type ExternalAmazonBedrockConverseStandardRequestSystem struct {
	Text string `json:"text"`
}

type ExternalAmazonBedrockConverseStandardRequestToolConfig struct {
	ToolChoice ExternalAmazonBedrockConverseStandardRequestToolConfigToolChoice `json:"toolChoice"`
	Tools      []ExternalAmazonBedrockConverseStandardRequestToolConfigTool     `json:"tools"`
}

type ExternalAmazonBedrockConverseStandardRequestToolConfigToolChoice struct {
	Auto *ExternalAmazonBedrockConverseStandardRequestToolConfigToolChoiceAuto `json:"auto,omitempty"`
}

type ExternalAmazonBedrockConverseStandardRequestToolConfigToolChoiceAuto struct {
}

type ExternalAmazonBedrockConverseStandardRequestToolConfigTool struct {
	ToolSpec ExternalAmazonBedrockConverseStandardRequestToolConfigToolToolSpec `json:"toolSpec"`
}

type ExternalAmazonBedrockConverseStandardRequestToolConfigToolToolSpec struct {
	Name        string                                                                        `json:"name"`
	Description string                                                                        `json:"description"`
	InputSchema ExternalAmazonBedrockConverseStandardRequestToolConfigToolToolSpecInputSchema `json:"inputSchema"`
}

type ExternalAmazonBedrockConverseStandardRequestToolConfigToolToolSpecInputSchema struct {
	Json ExternalAmazonBedrockConverseStandardRequestToolConfigToolToolSpecInputSchemaJson `json:"json"`
}

type ExternalAmazonBedrockConverseStandardRequestToolConfigToolToolSpecInputSchemaJson struct {
	Type                 string                            `json:"type"`
	Properties           ExternalToolInputSchemaProperties `json:"properties"`
	Required             []string                          `json:"required"`
	AdditionalProperties bool                              `json:"additionalProperties"`
}

type ExternalAmazonBedrockConverseStandardRequestInferenceConfig struct {
	MaxTokens int `json:"maxTokens"`
}

func NewExternalAmazonBedrockConverseStandardRequest(
	model string,
	messages []ExternalAmazonBedrockConverseStandardMessage,
	system []ExternalAmazonBedrockConverseStandardRequestSystem,
) *ExternalAmazonBedrockConverseStandardRequest {
	return &ExternalAmazonBedrockConverseStandardRequest{
		ModelId:  model,
		System:   system,
		Messages: messages,
		InferenceConfig: &ExternalAmazonBedrockConverseStandardRequestInferenceConfig{
			MaxTokens: 32000,
		},
	}
}

func (r *ExternalAmazonBedrockConverseStandardRequest) AddDefaultToolConfigIfNil() {
	if r.ToolConfig == nil {
		r.ToolConfig = &ExternalAmazonBedrockConverseStandardRequestToolConfig{
			Tools: make([]ExternalAmazonBedrockConverseStandardRequestToolConfigTool, 0),
			ToolChoice: ExternalAmazonBedrockConverseStandardRequestToolConfigToolChoice{
				Auto: &ExternalAmazonBedrockConverseStandardRequestToolConfigToolChoiceAuto{},
			},
		}
	}
}

func (r *ExternalAmazonBedrockConverseStandardRequest) AddMcpConnectionAsFunctionCallTools(
	mcpConnection *models.MCPConnection,
	relatedTools []*models.Tool,
) {
	if len(relatedTools) == 0 {
		return
	}

	r.AddDefaultToolConfigIfNil()

	for _, mcpTool := range relatedTools {
		r.appendFunctionCallTool(utils.McpToolToFunctionName(mcpConnection, mcpTool), mcpTool)
	}
}

func (r *ExternalAmazonBedrockConverseStandardRequest) AddSelfToolsAsFunctionCallTools(selfTools []*models.Tool) {
	if len(selfTools) == 0 {
		return
	}

	r.AddDefaultToolConfigIfNil()

	for _, selfTool := range selfTools {
		r.appendFunctionCallTool(utils.SelfToolToFunctionName(selfTool), selfTool)
	}
}

func (r *ExternalAmazonBedrockConverseStandardRequest) appendFunctionCallTool(toolName string, tool *models.Tool) {
	functionTool := ExternalAmazonBedrockConverseStandardRequestToolConfigToolToolSpec{
		Name:        toolName,
		Description: tool.Description,
		InputSchema: ExternalAmazonBedrockConverseStandardRequestToolConfigToolToolSpecInputSchema{
			Json: ExternalAmazonBedrockConverseStandardRequestToolConfigToolToolSpecInputSchemaJson{
				Type:                 tool.InputScheme.Type,
				Required:             tool.InputScheme.Required,
				AdditionalProperties: tool.InputScheme.AdditionalProperties,

				Properties: NewExternalToolInputSchemaPropertiesFromInternalModels(tool.InputScheme.Properties),
			},
		},
	}

	if functionTool.InputSchema.Json.Required == nil {
		functionTool.InputSchema.Json.Required = []string{}
	}

	r.ToolConfig.Tools = append(r.ToolConfig.Tools, ExternalAmazonBedrockConverseStandardRequestToolConfigTool{
		ToolSpec: functionTool,
	})
}
