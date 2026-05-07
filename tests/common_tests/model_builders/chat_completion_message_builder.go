package model_builders

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/go-faker/faker/v4"
)

type ChatCompletionMessageBuilder struct {
	role    string
	content string
}

func NewChatCompletionMessageBuilder() *ChatCompletionMessageBuilder {
	role := "user"
	content := faker.Sentence()

	return &ChatCompletionMessageBuilder{
		role:    role,
		content: content,
	}
}

func (b *ChatCompletionMessageBuilder) WithRole(role string) *ChatCompletionMessageBuilder {
	b.role = role
	return b
}

func (b *ChatCompletionMessageBuilder) WithContent(content string) *ChatCompletionMessageBuilder {
	b.content = content
	return b
}

func (b *ChatCompletionMessageBuilder) Build() *interfaces.ChatCompletionMessage {
	return &interfaces.ChatCompletionMessage{
		Role:    b.role,
		Content: b.content,
	}
}
