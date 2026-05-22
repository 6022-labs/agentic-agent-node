package services

import (
	"context"
	"errors"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/ethereum/go-ethereum/common"
)

type NftOwnerAuthorizer struct {
	identityResolver    common_interfaces.AgentIdentityResolverInterface
	collectionRequester identity_interfaces.AgentCollectionRequesterInterface
}

func NewNftOwnerAuthorizer(
	identityResolver common_interfaces.AgentIdentityResolverInterface,
	collectionRequester identity_interfaces.AgentCollectionRequesterInterface,
) *NftOwnerAuthorizer {
	return &NftOwnerAuthorizer{
		identityResolver:    identityResolver,
		collectionRequester: collectionRequester,
	}
}

func (a *NftOwnerAuthorizer) IsAuthorized(ctx context.Context, address common.Address) (bool, error) {
	identity, err := a.identityResolver.Resolve(ctx)
	if err != nil {
		// No NFT yet: open login while unclaimed.
		if errors.Is(err, common_interfaces.ErrIdentityNotFound) {
			return true, nil
		}
		return false, err
	}
	owner, err := a.collectionRequester.OwnerOf(common.HexToAddress(identity.CollectionAddress), identity.TokenId)
	if err != nil {
		return false, err
	}
	return owner == address, nil
}
