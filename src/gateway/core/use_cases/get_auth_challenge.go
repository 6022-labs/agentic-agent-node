package use_cases

import (
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/responses"
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
		return nil, errors.NewInternalError("challenge_failed", err.Error())
	}
	return &responses.GetAuthChallengeResponse{Nonce: nonce, ExpiresAt: expiresAt}, nil
}
