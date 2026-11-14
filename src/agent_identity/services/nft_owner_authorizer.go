package services

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
)

type NftOwnerAuthorizer struct {
	agentResolver       identity_interfaces.AgentResolverInterface
	stateRepository     repositories.AgentRegistrationStateRepositoryInterface
	collectionRequester identity_interfaces.AgentCollectionRequesterInterface
}

func NewNftOwnerAuthorizer(
	agentResolver identity_interfaces.AgentResolverInterface,
	stateRepository repositories.AgentRegistrationStateRepositoryInterface,
	collectionRequester identity_interfaces.AgentCollectionRequesterInterface,
) *NftOwnerAuthorizer {
	return &NftOwnerAuthorizer{
		agentResolver:       agentResolver,
		stateRepository:     stateRepository,
		collectionRequester: collectionRequester,
	}
}

func (a *NftOwnerAuthorizer) IsAuthorized(ctx context.Context, address common.Address) (bool, error) {
	agent, err := a.agentResolver.Get(ctx)
	if err != nil {
		return false, err
	}
	if agent == nil || agent.CollectionTokenId == nil {
		// No NFT yet: open login while unclaimed.
		return true, nil
	}

	state, err := a.stateRepository.GetByAgentId(ctx, agent.Id)
	if err != nil {
		return false, err
	}
	if !state.IsMinted() {
		// No NFT yet: open login while unclaimed.
		return true, nil
	}

	owner, err := a.collectionRequester.OwnerOf(
		lo.FromPtr(agent.ChainId),
		common.HexToAddress(agent.CollectionAddress),
		*agent.CollectionTokenId,
	)
	if err != nil {
		return false, err
	}
	return owner == address, nil
}
