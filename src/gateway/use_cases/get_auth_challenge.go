package use_cases

import (
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
)

type GetAuthChallenge struct {
	nonceStore gateway_interfaces.NonceStoreInterface
}

func NewGetAuthChallenge(nonceStore gateway_interfaces.NonceStoreInterface) *GetAuthChallenge {
	return &GetAuthChallenge{nonceStore: nonceStore}
}

func (u *GetAuthChallenge) Execute() (*responses.GetAuthChallengeResponse, error) {
	nonce, expiresAt, err := u.nonceStore.Generate()
	if err != nil {
		return nil, err
	}
	return &responses.GetAuthChallengeResponse{Nonce: nonce, ExpiresAt: expiresAt}, nil
}
