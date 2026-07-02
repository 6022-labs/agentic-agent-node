package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases/responses"
)

type GetAgentRegistration struct {
	client gateway_interfaces.AgentRegistrationClientInterface
}

func NewGetAgentRegistration(client gateway_interfaces.AgentRegistrationClientInterface) *GetAgentRegistration {
	return &GetAgentRegistration{client: client}
}

func (u *GetAgentRegistration) Execute(ctx context.Context) (*responses.AgentRegistrationResponse, error) {
	payload, err := u.client.Get(ctx)
	if err != nil {
		return nil, err
	}
	return u.toResponse(payload), nil
}

func (u *GetAgentRegistration) toResponse(p *models.ExternalAgentRegistrationPayload) *responses.AgentRegistrationResponse {
	return &responses.AgentRegistrationResponse{
		ChainId:           p.ChainId,
		CollectionAddress: p.CollectionAddress,
		TokenId:           p.TokenId,
		ProposalId:        p.ProposalId,
		Name:              p.Name,
		Role:              p.Role,
		CloneOf:           p.CloneOf,
		Owner:             p.Owner,
		Addresses:         p.Addresses,
		EnsDomain:         p.EnsDomain,
		Images:            p.Images,
		Attributes:        p.Attributes,
		MintStatus:        p.MintStatus,
		EnsStatus:         p.EnsStatus,
	}
}
