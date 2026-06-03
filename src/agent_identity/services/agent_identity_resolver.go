package services

import (
	"context"
	"errors"
	"sync"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/models"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// ErrIdentityNotFound is re-exported for backward compatibility; the canonical
// definition lives on the interface in common/interfaces.
var ErrIdentityNotFound = common_interfaces.ErrIdentityNotFound

type AgentIdentityResolver struct {
	logger              *zap.Logger
	collectionRequester identity_interfaces.AgentCollectionRequesterInterface
	managerRequester    identity_interfaces.AgentCollectionsManagerRequesterInterface

	mu    sync.Mutex
	cache *models.AgentIdentity
}

func NewAgentIdentityResolver(
	logger *zap.Logger,
	collectionRequester identity_interfaces.AgentCollectionRequesterInterface,
	managerRequester identity_interfaces.AgentCollectionsManagerRequesterInterface,
) *AgentIdentityResolver {
	return &AgentIdentityResolver{
		logger:              logger,
		collectionRequester: collectionRequester,
		managerRequester:    managerRequester,
	}
}

// agentWalletAddress returns the loaded agent wallet address.
//
// NOT IMPLEMENTED: bootstrap state holder removed. Rewire to vault-backed
// wallet storage before identity resolution works again.
func (r *AgentIdentityResolver) agentWalletAddress() (common.Address, error) {
	return common.Address{}, errors.New("not implemented: agent wallet address source removed with bootstrap state holder")
}

func (r *AgentIdentityResolver) Resolve(ctx context.Context) (*models.AgentIdentity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cache != nil {
		return r.cache, nil
	}

	walletAddr, err := r.agentWalletAddress()
	if err != nil {
		return nil, err
	}

	collections, err := r.managerRequester.ListAllCollections()
	if err != nil {
		return nil, err
	}

	for _, collectionAddr := range collections {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		tokenId, err := r.collectionRequester.TokenIdByWallet(collectionAddr, walletAddr)
		if err != nil {
			r.logger.Debug("walletToTokenId failed",
				zap.String("collection", collectionAddr.Hex()),
				zap.Error(err),
			)
			continue
		}
		if tokenId == 0 {
			continue
		}

		name, err := r.collectionRequester.NameOf(collectionAddr, tokenId)
		if err != nil {
			return nil, err
		}

		identity := &models.AgentIdentity{
			CollectionAddress: collectionAddr.Hex(),
			TokenId:           tokenId,
			Name:              name,
		}
		r.cache = identity
		r.logger.Info("agent identity resolved",
			zap.String("collection", identity.CollectionAddress),
			zap.Uint64("tokenId", identity.TokenId),
			zap.String("name", identity.Name),
		)
		return identity, nil
	}

	return nil, ErrIdentityNotFound
}
