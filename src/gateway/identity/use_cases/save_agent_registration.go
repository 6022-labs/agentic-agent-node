package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases/responses"
)

type SaveAgentRegistration struct {
	client gateway_interfaces.AgentRegistrationClientInterface
}

func NewSaveAgentRegistration(client gateway_interfaces.AgentRegistrationClientInterface) *SaveAgentRegistration {
	return &SaveAgentRegistration{client: client}
}

// Execute forwards the draft to the runtime, injecting the authenticated owner.
func (u *SaveAgentRegistration) Execute(ctx context.Context, req *requests.SaveAgentRegistrationRequest, owner string) (*responses.AgentRegistrationResponse, error) {
	if err := req.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	payload, err := u.client.Save(ctx, &models.ExternalSaveAgentRegistrationPayload{
		ChainId:           req.ChainId,
		CollectionAddress: req.CollectionAddress,
		Owner:             owner,
		Name:              req.Name,
		Role:              req.Role,
		CloneOf:           req.CloneOf,
		Image:             req.Image,
		Attributes:        req.Attributes,
	})
	if err != nil {
		return nil, err
	}
	return u.toResponse(payload), nil
}

func (u *SaveAgentRegistration) toResponse(p *models.ExternalAgentRegistrationPayload) *responses.AgentRegistrationResponse {
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
