package use_cases

import (
	"context"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type SignAgentCard struct {
	agentCardSigner identity_interfaces.AgentCardSignerInterface
}

func NewSignAgentCard(agentCardSigner identity_interfaces.AgentCardSignerInterface) *SignAgentCard {
	return &SignAgentCard{agentCardSigner: agentCardSigner}
}

func (u *SignAgentCard) Execute(
	ctx context.Context,
	request *requests.SignAgentCardRequest,
) (*responses.AgentCardSignatureResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	signature, err := u.agentCardSigner.Sign(ctx, request.Card)
	if err != nil {
		return nil, errors.NewInternalError("agent_card_signing_failed", err)
	}

	response := &responses.AgentCardSignatureResponse{}
	response.FromModel(signature)
	return response, nil
}
