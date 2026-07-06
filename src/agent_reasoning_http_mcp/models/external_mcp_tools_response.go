package models

import (
	reasoning_models "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

type ExternalMcpToolsResponse struct {
	Tools []ExternalMcpToolsResponseTool `json:"tools"`
}

type ExternalMcpToolsResponseTool struct {
	Name        string                              `json:"name"`
	Title       string                              `json:"title"`
	Description string                              `json:"description"`
	InputSchema ExternalMcpToolsResponseInputSchema `json:"inputSchema"`
}

type ExternalMcpToolsResponseInputSchema struct {
	Type                 string                                                 `json:"type"`
	Properties           map[string]ExternalMcpToolsResponseInputSchemaProperty `json:"properties"`
	Required             []string                                               `json:"required"`
	AdditionalProperties bool                                                   `json:"additionalProperties"`
	Schema               string                                                 `json:"$schema"`
}

type ExternalMcpToolsResponseInputSchemaProperty struct {
	Type        string                                       `json:"type"`
	Items       *ExternalMcpToolsResponseInputSchemaProperty `json:"items,omitempty"`
	Description *string                                      `json:"description,omitempty"`
}

func (e *ExternalMcpToolsResponse) ToFunctionTools() []*reasoning_models.Tool {
	tools := make([]*reasoning_models.Tool, len(e.Tools))
	for index, tool := range e.Tools {
		tools[index] = &reasoning_models.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputScheme: reasoning_models.ToolInputScheme{
				Type:                 tool.InputSchema.Type,
				Required:             tool.InputSchema.Required,
				AdditionalProperties: tool.InputSchema.AdditionalProperties,
				Schema:               tool.InputSchema.Schema,
			},
		}

		for propName, prop := range tool.InputSchema.Properties {
			if tools[index].InputScheme.Properties == nil {
				tools[index].InputScheme.Properties = make(map[string]reasoning_models.ToolInputSchemeProperty)
			}

			tools[index].InputScheme.Properties[propName] = prop.ToInternalModel()
		}
	}

	return tools
}

func (p ExternalMcpToolsResponseInputSchemaProperty) ToInternalModel() reasoning_models.ToolInputSchemeProperty {
	this := reasoning_models.ToolInputSchemeProperty{
		Type:        p.Type,
		Description: p.Description,
	}

	if p.Items != nil {
		internalItems := p.Items.ToInternalModel()
		this.Items = &internalItems
	}

	return this
}
