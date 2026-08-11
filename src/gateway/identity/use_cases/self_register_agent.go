package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases/responses"
)

type SelfRegisterAgent struct {
	client gateway_interfaces.AgentRegistrationClientInterface
}

func NewSelfRegisterAgent(client gateway_interfaces.AgentRegistrationClientInterface) *SelfRegisterAgent {
	return &SelfRegisterAgent{client: client}
}

func (u *SelfRegisterAgent) Execute(ctx context.Context) (*responses.RegistrationStatusResponse, error) {
	payload, err := u.client.Self(ctx)
	if err != nil {
		return nil, err
	}
	return u.toResponse(payload), nil
}

func (u *SelfRegisterAgent) toResponse(
	p *models.ExternalRegistrationStatusPayload,
) *responses.RegistrationStatusResponse {
	return &responses.RegistrationStatusResponse{
		MintStatus: p.MintStatus,
		EnsStatus:  p.EnsStatus,
		TokenId:    p.TokenId,
		ProposalId: p.ProposalId,
		MintTxHash: p.MintTxHash,
		EnsTxHash:  p.EnsTxHash,
	}
}
