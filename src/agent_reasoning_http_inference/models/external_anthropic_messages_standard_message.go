package models

type ExternalAnthropicMessagesStandardMessage struct {
	// Common fields
	Role       string                                            `json:"role"`
	Content    []ExternalAnthropicMessagesStandardMessageContent `json:"content,omitempty"`
	StopReason *string                                           `json:"stop_reason,omitempty"`
}

type ExternalAnthropicMessagesStandardMessageContent struct {
	// Common fields
	Type string `json:"type"`

	// `text` type fields
	Text *string `json:"text,omitempty"`

	// `thinking` type fields
	Signature *string `json:"signature,omitempty"`
	Thinking  *string `json:"thinking,omitempty"`

	// `redacted_thinking` type fields
	Data *string `json:"data,omitempty"`

	// {tool,mcp_tool,server_tool}_use fields
	Id    *string        `json:"id,omitempty"`
	Name  *string        `json:"name,omitempty"`
	Input map[string]any `json:"input,omitempty"`

	// *_tool_result fields (tool, mcp, web_fetch, web_search, code_execution,
	// bash_code_execution, text_editor_code_execution)
	Content   any     `json:"content,omitempty"`
	ToolUseId *string `json:"tool_use_id,omitempty"`

	// `mcp_tool_use` fields
	ServerName *string `json:"server_name,omitempty"`

	// `container_upload` fields
	FileId *string `json:"file_id,omitempty"`
}
