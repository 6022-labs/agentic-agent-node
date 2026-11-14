package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
)

// GetAgentMintInfo assembles the mint() calldata + moderator state so the connected wallet can submit the tx itself.
type GetAgentMintInfo struct {
	agentResolver            identity_interfaces.AgentResolverInterface
	agentAddressRepository   repositories.AgentAddressRepositoryInterface
	agentCollectionRequester identity_interfaces.AgentCollectionRequesterInterface
}

func NewGetAgentMintInfo(
	agentResolver identity_interfaces.AgentResolverInterface,
	agentAddressRepository repositories.AgentAddressRepositoryInterface,
	agentCollectionRequester identity_interfaces.AgentCollectionRequesterInterface,
) *GetAgentMintInfo {
	return &GetAgentMintInfo{
		agentResolver:            agentResolver,
		agentAddressRepository:   agentAddressRepository,
		agentCollectionRequester: agentCollectionRequester,
	}
}

func (u *GetAgentMintInfo) Execute(ctx context.Context) (*responses.MintInfoResponse, error) {
	agent, err := u.agentResolver.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("agent_lookup_failed", err)
	}
	if agent == nil {
		return nil, errors.NewNotFoundError("draft_not_found", "no agent registration has been saved yet")
	}
	if agent.Owner == nil {
		return nil, errors.NewConflictError("owner_missing", "agent draft has no owner")
	}

	addresses, err := u.agentAddressRepository.GetAllByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("agent_addresses_lookup_failed", err)
	}

	collection := common.HexToAddress(agent.CollectionAddress)
	owner := common.HexToAddress(*agent.Owner)
	chainId := lo.FromPtr(agent.ChainId)

	moderatorCount, err := u.agentCollectionRequester.ModeratorCount(chainId, collection)
	if err != nil {
		return nil, errors.NewUnavailableError(
			"collection_read_failed",
			"failed to read the collection moderators",
			err,
		)
	}
	ownerIsModerator := false
	if moderatorCount > 0 {
		ownerIsModerator, err = u.agentCollectionRequester.IsModerator(chainId, collection, owner)
		if err != nil {
			return nil, errors.NewUnavailableError(
				"collection_read_failed",
				"failed to read the collection moderators",
				err,
			)
		}
	}

	return &responses.MintInfoResponse{
		CollectionAddress: collection.Hex(),
		To:                owner.Hex(),
		Name:              lo.FromPtr(agent.Name),
		Addresses:         addresses,
		Images:            responses.KeyValue{}.FromMap(agent.Images),
		Attributes:        responses.KeyValue{}.FromMap(agent.Attributes),
		CloneOf:           agent.CloneOf,
		ModeratorCount:    moderatorCount,
		OwnerIsModerator:  ownerIsModerator,
	}, nil
}
