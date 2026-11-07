package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/responses"
)

type SetAgentGuidance struct {
	agentGuidanceClient gateway_interfaces.AgentGuidanceClientInterface
}

func NewSetAgentGuidance(agentGuidanceClient gateway_interfaces.AgentGuidanceClientInterface) *SetAgentGuidance {
	return &SetAgentGuidance{agentGuidanceClient: agentGuidanceClient}
}

func (u *SetAgentGuidance) Execute(
	ctx context.Context,
	request *requests.SetAgentGuidanceRequest,
) (*responses.AgentGuidanceResponse, error) {
	payload, err := u.agentGuidanceClient.Set(ctx, &models.ExternalAgentGuidancePayload{Content: request.Content})
	if err != nil {
		return nil, err
	}
	return &responses.AgentGuidanceResponse{Content: payload.Content}, nil
}
