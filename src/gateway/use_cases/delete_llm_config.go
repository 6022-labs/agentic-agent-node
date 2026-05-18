package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
)

type DeleteLlmConfig struct {
	client gateway_interfaces.LLMConfigClientInterface
}

func NewDeleteLlmConfig(client gateway_interfaces.LLMConfigClientInterface) *DeleteLlmConfig {
	return &DeleteLlmConfig{client: client}
}

func (u *DeleteLlmConfig) Execute(ctx context.Context) *usecase_result.Result {
	if err := u.client.Delete(ctx); err != nil {
		return usecase_result.Unavailable("runtime_unavailable", err.Error())
	}
	return usecase_result.NoContent()
}
