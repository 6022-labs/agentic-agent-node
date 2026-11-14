package use_cases

import (
	"context"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/samber/lo"
)

type GetAgentSoul struct {
	agentResolver identity_interfaces.AgentResolverInterface
}

func NewGetAgentSoul(agentResolver identity_interfaces.AgentResolverInterface) *GetAgentSoul {
	return &GetAgentSoul{agentResolver: agentResolver}
}

func (u *GetAgentSoul) Execute(ctx context.Context) (*responses.AgentSoulResponse, error) {
	agent, err := u.agentResolver.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("soul_lookup_failed", err)
	}
	if agent == nil {
		return nil, errors.NewNotFoundError("identity_not_found", "agent identity not found")
	}
	return &responses.AgentSoulResponse{Soul: lo.FromPtr(agent.Soul)}, nil
}
