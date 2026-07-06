package models

import (
	reasoning_models "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

type ExternalToolInputSchemaProperty struct {
	Type                 string                           `json:"type"`
	Items                *ExternalToolInputSchemaProperty `json:"items,omitempty"`
	Description          *string                          `json:"description,omitempty"`
	AdditionalProperties *bool                            `json:"additionalProperties,omitempty"`
}

type ExternalToolInputSchemaProperties map[string]ExternalToolInputSchemaProperty

func NewExternalToolInputSchemaPropertyFromInternalModel(internalProperty reasoning_models.ToolInputSchemeProperty) ExternalToolInputSchemaProperty {
	this := ExternalToolInputSchemaProperty{
		Type:                 internalProperty.Type,
		Description:          internalProperty.Description,
		AdditionalProperties: internalProperty.AdditionalProperties,
	}

	if internalProperty.Items != nil {
		internalItems := NewExternalToolInputSchemaPropertyFromInternalModel(*internalProperty.Items)
		this.Items = &internalItems
	}

	return this
}

func NewExternalToolInputSchemaPropertiesFromInternalModels(internalProperties map[string]reasoning_models.ToolInputSchemeProperty) ExternalToolInputSchemaProperties {
	externalProperties := make(ExternalToolInputSchemaProperties)

	for key, internalProperty := range internalProperties {
		externalProperties[key] = NewExternalToolInputSchemaPropertyFromInternalModel(internalProperty)
	}

	return externalProperties
}

func (p *ExternalToolInputSchemaProperty) ToInternalModel() reasoning_models.ToolInputSchemeProperty {
	internalModel := reasoning_models.ToolInputSchemeProperty{
		Type:                 p.Type,
		Description:          p.Description,
		AdditionalProperties: p.AdditionalProperties,
	}

	if p.Items != nil {
		internalItems := p.Items.ToInternalModel()
		internalModel.Items = &internalItems
	}

	return internalModel
}

func (p ExternalToolInputSchemaProperties) ToInternalModels() map[string]reasoning_models.ToolInputSchemeProperty {
	internalProperties := make(map[string]reasoning_models.ToolInputSchemeProperty)

	for key, externalProperty := range p {
		internalProperties[key] = externalProperty.ToInternalModel()
	}

	return internalProperties
}
