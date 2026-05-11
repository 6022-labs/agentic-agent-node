package models

type ExternalOllamaChatMessageStandardMessage struct {
	Role      string                                             `json:"role"`
	Content   string                                             `json:"content"`
	ToolName  *string                                            `json:"tool_name,omitempty"`
	ToolCalls []ExternalOllamaChatMessageStandardMessageToolCall `json:"tool_calls,omitempty"`
}

type ExternalOllamaChatMessageStandardMessageToolCall struct {
	Function *ExternalOllamaChatMessageStandardMessageToolCallFunction `json:"function,omitempty"`
}

type ExternalOllamaChatMessageStandardMessageToolCallFunction struct {
	Name        string                                                             `json:"name"`
	Index       *int                                                               `json:"index,omitempty"`
	Arguments   map[string]any                                                     `json:"arguments,omitempty"`
	Parameters  *ExternalOllamaChatMessageStandardMessageToolCallFunctionArguments `json:"parameters,omitempty"`
	Description *string                                                            `json:"description,omitempty"`
}

type ExternalOllamaChatMessageStandardMessageToolCallFunctionArguments struct {
	Type       string                            `json:"type"`
	Required   []string                          `json:"required"`
	Properties ExternalToolInputSchemaProperties `json:"properties"`
}

