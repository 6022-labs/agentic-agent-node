package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
)

type CheckAgentMint struct {
	client gateway_interfaces.AgentRegistrationClientInterface
}

func NewCheckAgentMint(client gateway_interfaces.AgentRegistrationClientInterface) *CheckAgentMint {
	return &CheckAgentMint{client: client}
}

func (u *CheckAgentMint) Execute(ctx context.Context) (*responses.RegistrationStatusResponse, error) {
	payload, err := u.client.CheckMint(ctx)
	if err != nil {
		return nil, err
	}
	return u.toResponse(payload), nil
}

func (u *CheckAgentMint) toResponse(p *models.ExternalRegistrationStatusPayload) *responses.RegistrationStatusResponse {
	return &responses.RegistrationStatusResponse{
		MintStatus: p.MintStatus,
		EnsStatus:  p.EnsStatus,
		TokenId:    p.TokenId,
		ProposalId: p.ProposalId,
		MintTxHash: p.MintTxHash,
		EnsTxHash:  p.EnsTxHash,
	}
}
