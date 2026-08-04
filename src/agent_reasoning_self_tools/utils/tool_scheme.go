package utils

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

func EmptyObjectScheme() models.ToolInputScheme {
	return ObjectScheme(map[string]models.ToolInputSchemeProperty{}, nil)
}

func ObjectScheme(
	properties map[string]models.ToolInputSchemeProperty,
	required []string,
) models.ToolInputScheme {
	return models.ToolInputScheme{
		Type:       "object",
		Properties: properties,
		Required:   required,
	}
}

func StringProperty(description string) models.ToolInputSchemeProperty {
	return models.ToolInputSchemeProperty{
		Type:        "string",
		Description: &description,
	}
}

func IntegerProperty(description string) models.ToolInputSchemeProperty {
	return models.ToolInputSchemeProperty{
		Type:        "integer",
		Description: &description,
	}
}

func BooleanProperty(description string) models.ToolInputSchemeProperty {
	return models.ToolInputSchemeProperty{
		Type:        "boolean",
		Description: &description,
	}
}

func ObjectProperty(description string) models.ToolInputSchemeProperty {
	additionalProperties := true
	return models.ToolInputSchemeProperty{
		Type:                 "object",
		Description:          &description,
		AdditionalProperties: &additionalProperties,
	}
}
