package models

// Tool is a provider-facing tool definition, whatever its origin.
type Tool struct {
	Name        string
	Description string
	InputScheme ToolInputScheme
}

type ToolInputScheme struct {
	Type                 string
	Properties           map[string]ToolInputSchemeProperty
	Required             []string
	AdditionalProperties bool
	Schema               string
}

type ToolInputSchemeProperty struct {
	Type                 string
	Items                *ToolInputSchemeProperty
	Description          *string
	AdditionalProperties *bool
}
