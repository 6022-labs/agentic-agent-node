package use_cases

import (
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/responses"
	"github.com/ethereum/go-ethereum/common"
)

type GetAuthChallenge struct {
	nonceStore         gateway_interfaces.NonceStoreInterface
	siweMessageBuilder gateway_interfaces.SiweMessageBuilderInterface
}

func NewGetAuthChallenge(
	nonceStore gateway_interfaces.NonceStoreInterface,
	siweMessageBuilder gateway_interfaces.SiweMessageBuilderInterface,
) *GetAuthChallenge {
	return &GetAuthChallenge{
		nonceStore:         nonceStore,
		siweMessageBuilder: siweMessageBuilder,
	}
}

func (u *GetAuthChallenge) Execute(request *requests.GetAuthChallengeRequest) (*responses.GetAuthChallengeResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	address := common.HexToAddress(request.Address)
	entry, err := u.nonceStore.Generate(address, request.ChainId)
	if err != nil {
		return nil, errors.NewInternalError("challenge_failed", err.Error())
	}

	message := u.siweMessageBuilder.Build(address, entry.ChainId, entry.Nonce, entry.IssuedAt, entry.ExpiresAt)

	return &responses.GetAuthChallengeResponse{
		Message:   message,
		ExpiresAt: entry.ExpiresAt,
	}, nil
}
