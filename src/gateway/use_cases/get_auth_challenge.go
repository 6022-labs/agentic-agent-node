package use_cases

import (
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
)

type GetAuthChallenge struct {
	nonceStore gateway_interfaces.NonceStoreInterface
}

func NewGetAuthChallenge(nonceStore gateway_interfaces.NonceStoreInterface) *GetAuthChallenge {
	return &GetAuthChallenge{nonceStore: nonceStore}
}

func (u *GetAuthChallenge) Execute() *usecase_result.Result {
	nonce, expiresAt, err := u.nonceStore.Generate()
	if err != nil {
		return usecase_result.InternalError("challenge_failed", err.Error())
	}
	return usecase_result.Ok(&responses.GetAuthChallengeResponse{Nonce: nonce, ExpiresAt: expiresAt})
}
