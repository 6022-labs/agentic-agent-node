package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
)

type DeleteLlmConfig struct {
	client gateway_interfaces.LLMConfigClientInterface
}

func NewDeleteLlmConfig(client gateway_interfaces.LLMConfigClientInterface) *DeleteLlmConfig {
	return &DeleteLlmConfig{client: client}
}

func (u *DeleteLlmConfig) Execute(ctx context.Context) error {
	return u.client.Delete(ctx)
}
