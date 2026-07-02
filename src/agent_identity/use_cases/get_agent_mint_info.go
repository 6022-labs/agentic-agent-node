package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/ethereum/go-ethereum/common"
)

// GetAgentMintInfo assembles the mint() calldata + moderator state so the connected wallet can submit the tx itself.
type GetAgentMintInfo struct {
	agentRepository          repositories.AgentRepositoryInterface
	agentCollectionRequester identity_interfaces.AgentCollectionRequesterInterface
}

func NewGetAgentMintInfo(
	agentRepository repositories.AgentRepositoryInterface,
	agentCollectionRequester identity_interfaces.AgentCollectionRequesterInterface,
) *GetAgentMintInfo {
	return &GetAgentMintInfo{agentRepository: agentRepository, agentCollectionRequester: agentCollectionRequester}
}

func (u *GetAgentMintInfo) Execute(ctx context.Context) (*responses.MintInfoResponse, error) {
	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("agent_lookup_failed", err.Error())
	}
	if agent == nil {
		return nil, errors.NewNotFoundError("draft_not_found", "no agent registration has been saved yet")
	}
	if agent.Owner == nil {
		return nil, errors.NewConflictError("owner_missing", "agent draft has no owner")
	}

	collection := common.HexToAddress(agent.CollectionAddress)
	owner := common.HexToAddress(*agent.Owner)

	moderatorCount, err := u.agentCollectionRequester.ModeratorCount(agent.ChainId, collection)
	if err != nil {
		return nil, errors.NewUnavailableError("collection_read_failed", err.Error())
	}
	ownerIsModerator := false
	if moderatorCount > 0 {
		ownerIsModerator, err = u.agentCollectionRequester.IsModerator(agent.ChainId, collection, owner)
		if err != nil {
			return nil, errors.NewUnavailableError("collection_read_failed", err.Error())
		}
	}

	return &responses.MintInfoResponse{
		CollectionAddress: collection.Hex(),
		To:                owner.Hex(),
		Role:              agent.Role.String(),
		Name:              agent.Name,
		Addresses:         agent.Addresses,
		Images:            responses.KeyValue{}.FromMap(agent.Images),
		Attributes:        responses.KeyValue{}.FromMap(agent.Attributes),
		CloneOf:           agent.CloneOf,
		ModeratorCount:    moderatorCount,
		OwnerIsModerator:  ownerIsModerator,
	}, nil
}
