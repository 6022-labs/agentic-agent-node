package utils

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
)

type conversationMessageHashFields struct {
	Role           string `json:"role"`
	Content        string `json:"content"`
	ToolCalls      string `json:"toolCalls"`
	ToolCallId     string `json:"toolCallId"`
	ConversationId string `json:"conversationId"`
}

// ComputeConversationMessageHash returns the SHA-512 dedup hash; demojize Content first so reasoning + knowledge agree.
func ComputeConversationMessageHash(role, content, toolCalls, toolCallId, conversationId string) (string, error) {
	b, err := json.Marshal(conversationMessageHashFields{
		Role:           role,
		Content:        content,
		ToolCalls:      toolCalls,
		ToolCallId:     toolCallId,
		ConversationId: conversationId,
	})
	if err != nil {
		return "", err
	}
	sum := sha512.Sum512(b)
	return hex.EncodeToString(sum[:]), nil
}
