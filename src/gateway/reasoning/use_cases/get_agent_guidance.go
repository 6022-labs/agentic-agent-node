package use_cases

import (
	"context"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/responses"
)

type GetAgentGuidance struct {
	agentGuidanceClient gateway_interfaces.AgentGuidanceClientInterface
}

func NewGetAgentGuidance(agentGuidanceClient gateway_interfaces.AgentGuidanceClientInterface) *GetAgentGuidance {
	return &GetAgentGuidance{agentGuidanceClient: agentGuidanceClient}
}

func (u *GetAgentGuidance) Execute(ctx context.Context) (*responses.AgentGuidanceResponse, error) {
	payload, err := u.agentGuidanceClient.Get(ctx)
	if err != nil {
		return nil, apperrors.NewUnavailableError("runtime_unavailable", err.Error())
	}
	return &responses.AgentGuidanceResponse{Content: payload.Content}, nil
}
