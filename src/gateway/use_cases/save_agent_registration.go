package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
)

type SaveAgentRegistration struct {
	client gateway_interfaces.AgentRegistrationClientInterface
}

func NewSaveAgentRegistration(client gateway_interfaces.AgentRegistrationClientInterface) *SaveAgentRegistration {
	return &SaveAgentRegistration{client: client}
}

// Execute forwards the draft to the runtime, injecting the authenticated owner.
func (u *SaveAgentRegistration) Execute(ctx context.Context, req *requests.SaveAgentRegistrationRequest, owner string) (*responses.AgentRegistrationResponse, error) {
	payload, err := u.client.Save(ctx, &runtime_models.ExternalSaveAgentRegistrationPayload{
		CollectionAddress: req.CollectionAddress,
		Owner:             owner,
		Name:              req.Name,
		Role:              req.Role,
		CloneOf:           req.CloneOf,
		Images:            req.Images,
		Attributes:        req.Attributes,
	})
	if err != nil {
		return nil, err
	}
	return u.toResponse(payload), nil
}

func (u *SaveAgentRegistration) toResponse(p *runtime_models.ExternalAgentRegistrationPayload) *responses.AgentRegistrationResponse {
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
