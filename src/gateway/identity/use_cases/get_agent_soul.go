package use_cases

import (
	"context"
	"errors"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases/responses"
)

type GetAgentSoul struct {
	agentSoulClient identity_interfaces.AgentSoulClientInterface
}

func NewGetAgentSoul(agentSoulClient identity_interfaces.AgentSoulClientInterface) *GetAgentSoul {
	return &GetAgentSoul{agentSoulClient: agentSoulClient}
}

func (u *GetAgentSoul) Execute(ctx context.Context) (*responses.AgentSoulResponse, error) {
	payload, err := u.agentSoulClient.Get(ctx)
	if errors.Is(err, identity_interfaces.ErrAgentSoulNotFound) {
		return nil, apperrors.NewNotFoundError("identity_not_found", "agent identity not found")
	}
	if err != nil {
		return nil, err
	}
	return &responses.AgentSoulResponse{Soul: payload.Soul}, nil
}
