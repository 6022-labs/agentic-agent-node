package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
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

func (u *GetAgentRegistration) toResponse(p *runtime_models.ExternalAgentRegistrationPayload) *responses.AgentRegistrationResponse {
	return &responses.AgentRegistrationResponse{
		CollectionAddress: p.CollectionAddress,
		TokenId:           p.TokenId,
		ProposalId:        p.ProposalId,
		Name:              p.Name,
		Role:              p.Role,
		CloneOf:           p.CloneOf,
		Owner:             p.Owner,
		WalletAddress:     p.WalletAddress,
		EnsDomain:         p.EnsDomain,
		Images:            p.Images,
		Attributes:        p.Attributes,
		MintStatus:        p.MintStatus,
		EnsStatus:         p.EnsStatus,
	}
}
