package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
)

type DeleteLlmConfig struct {
	client gateway_interfaces.LLMConfigClientInterface
}

func NewDeleteLlmConfig(client gateway_interfaces.LLMConfigClientInterface) *DeleteLlmConfig {
	return &DeleteLlmConfig{client: client}
}

func (u *DeleteLlmConfig) Execute(ctx context.Context) error {
	if err := u.client.Delete(ctx); err != nil {
		return errors.NewUnavailableError("runtime_unavailable", err.Error())
	}
	return nil
}
