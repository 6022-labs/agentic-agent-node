package services

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// LLMConnectionValidator probes the provider so validate and setup reject unusable credentials alike.
type LLMConnectionValidator struct {
	chatCompletionRequester interfaces.ChatCompletionRequesterInterface
}

func NewLLMConnectionValidator(
	chatCompletionRequester interfaces.ChatCompletionRequesterInterface,
) *LLMConnectionValidator {
	return &LLMConnectionValidator{chatCompletionRequester: chatCompletionRequester}
}

func (v *LLMConnectionValidator) Validate(
	ctx context.Context,
	llmConnection *models.LLMConnection,
	apiKey string,
) error {
	if err := v.chatCompletionRequester.ProbeConnection(ctx, llmConnection, apiKey); err != nil {
		return errors.NewValidationError("connection", err.Error())
	}
	return nil
}
