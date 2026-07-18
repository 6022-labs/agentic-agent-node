package services

import (
	"context"
	"fmt"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

// AgentChainRecoverer reconciles the on-chain identity of the node's existing,
// not-yet-minted agent using the wallet scoped to its id. It only ever touches the
// agent already persisted in the DB — it never enumerates the secret store, so a
// shared Vault can't leak another deployment's wallet into this node.
type AgentChainRecoverer struct {
	logger                           *zap.Logger
	walletCredentialsStorage         identity_interfaces.WalletCredentialsStorageInterface
	agentCollectionsManagerRequester identity_interfaces.AgentCollectionsManagerRequesterInterface
	agentCollectionRequester         identity_interfaces.AgentCollectionRequesterInterface
	agentRepository                  repositories.AgentRepositoryInterface
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface
	registrationReconciler           identity_interfaces.RegistrationReconcilerInterface
	ensRegistrar                     identity_interfaces.EnsRegistrarInterface
	ensDomainBuilder                 identity_interfaces.EnsDomainBuilderInterface
	agentCollectionsManagersSettings *runtime_common_settings.AgentCollectionsManagersSettings
}

func NewAgentChainRecoverer(
	logger *zap.Logger,
	walletCredentialsStorage identity_interfaces.WalletCredentialsStorageInterface,
	agentCollectionsManagerRequester identity_interfaces.AgentCollectionsManagerRequesterInterface,
	agentCollectionRequester identity_interfaces.AgentCollectionRequesterInterface,
	agentRepository repositories.AgentRepositoryInterface,
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface,
	registrationReconciler identity_interfaces.RegistrationReconcilerInterface,
	ensRegistrar identity_interfaces.EnsRegistrarInterface,
	ensDomainBuilder identity_interfaces.EnsDomainBuilderInterface,
	agentCollectionsManagersSettings *runtime_common_settings.AgentCollectionsManagersSettings,
) *AgentChainRecoverer {
	return &AgentChainRecoverer{
		logger:                           logger,
		walletCredentialsStorage:         walletCredentialsStorage,
		agentCollectionsManagerRequester: agentCollectionsManagerRequester,
		agentCollectionRequester:         agentCollectionRequester,
		agentRepository:                  agentRepository,
		agentRegistrationStateRepository: agentRegistrationStateRepository,
		registrationReconciler:           registrationReconciler,
		ensRegistrar:                     ensRegistrar,
		ensDomainBuilder:                 ensDomainBuilder,
		agentCollectionsManagersSettings: agentCollectionsManagersSettings,
	}
}

func (s *AgentChainRecoverer) Recover(ctx context.Context) (bool, error) {
	agent, err := s.agentRepository.Get(ctx)
	if err != nil {
		return false, err
	}
	if agent == nil {
		return false, nil
	}

	cred, found, err := s.walletCredentialsStorage.GetForId(ctx, agent.Id.String(), enums.AddressTypeEvm)
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}

	collection, tokenId, chainId, ok := s.findOnChain(common.HexToAddress(cred.Address))
	if !ok {
		return false, nil
	}

	if err := s.recoverMintedAgent(ctx, agent, collection, tokenId, chainId); err != nil {
		return false, err
	}
	return true, nil
}

// findOnChain scans every configured manager chain for an NFT the wallet controls.
func (s *AgentChainRecoverer) findOnChain(wallet common.Address) (common.Address, uint64, string, bool) {
	for _, chainId := range s.agentCollectionsManagersSettings.ChainIds() {
		collection, tokenId, _, err := s.agentCollectionsManagerRequester.FindAgentByAddress(chainId, enums.AddressTypeEvm, wallet.Hex())
		if err != nil {
			s.logger.Warn("scan collections failed during recovery",
				zap.String("chainId", chainId), zap.Error(err))
			continue
		}
		if tokenId > 0 {
			return collection, tokenId, chainId, true
		}
	}
	return common.Address{}, 0, "", false
}

func (s *AgentChainRecoverer) recoverMintedAgent(
	ctx context.Context,
	agent *models.Agent,
	collection common.Address,
	tokenId uint64,
	chainId string,
) error {
	info, err := s.agentCollectionRequester.InformationOf(chainId, collection, tokenId)
	if err != nil {
		return fmt.Errorf("read agent information: %w", err)
	}
	collectionName, err := s.agentCollectionRequester.CollectionName(chainId, collection)
	if err != nil {
		return fmt.Errorf("read collection name: %w", err)
	}

	var role enums.AgentRole
	_ = role.Set(info.Role)

	agent.ChainId = lo.ToPtr(chainId)
	agent.CollectionAddress = collection.Hex()
	agent.Name = lo.ToPtr(info.Name)
	agent.CloneOf = info.CloneOf
	agent.EnsDomain = lo.ToPtr(s.ensDomainBuilder.Build(info.Name, collectionName, chainId))
	agent.Role = &role
	agent.Images = info.Images
	agent.Attributes = info.Attributes

	state, err := s.agentRegistrationStateRepository.GetByAgentId(ctx, agent.Id)
	if err != nil {
		return fmt.Errorf("load registration state: %w", err)
	}

	// ReconcileMint fills creator/owner/token and marks mint confirmed, persisting the agent.
	_, savedState, err := s.registrationReconciler.ReconcileMint(ctx, agent, state, collection, tokenId, nil)
	if err != nil {
		return fmt.Errorf("reconcile mint: %w", err)
	}

	provisioned, err := s.ensRegistrar.Provisioned(ctx, chainId, collection, info.Name)
	if err != nil {
		s.logger.Warn("ENS provisioned check failed during recovery; leaving ENS status pending", zap.Error(err))
	} else if provisioned {
		if _, err := s.registrationReconciler.ReconcileEnsRegistry(ctx, agent.Id, savedState); err != nil {
			return fmt.Errorf("reconcile ens registry: %w", err)
		}
	}

	s.logger.Info("recovered agent identity from chain",
		zap.String("agentId", agent.Id.String()),
		zap.String("chainId", chainId),
		zap.String("collection", collection.Hex()),
		zap.Uint64("tokenId", tokenId),
		zap.String("name", info.Name),
	)
	return nil
}
