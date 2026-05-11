package utils

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
)

type conversationMessageHashFields struct {
	Role           string `json:"role"`
	Author         string `json:"author"`
	Content        string `json:"content"`
	Platform       string `json:"platform"`
	ConversationId string `json:"conversationId"`
}

// ComputeConversationMessageHash returns the canonical SHA-512 hash. Caller
// must demojize Content first so reasoning + knowledge agree.
func ComputeConversationMessageHash(role, author, platform, content, conversationId string) (string, error) {
	b, err := json.Marshal(conversationMessageHashFields{
		Role:           role,
		Author:         author,
		Content:        content,
		Platform:       platform,
		ConversationId: conversationId,
	})
	if err != nil {
		return "", err
	}
	sum := sha512.Sum512(b)
	return hex.EncodeToString(sum[:]), nil
}
