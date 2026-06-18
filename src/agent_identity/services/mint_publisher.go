package services

import (
	"context"
	"fmt"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/ethereum/go-ethereum/common"
)

// MintPublisher mints the agent's NFT with the node wallet (the self path): it
// inspects the collection's moderators to choose a direct mint or a mint
// proposal, then reconciles the rows. Paired with EnsPublisher.
type MintPublisher struct {
	agentCollectionRequester         identity_interfaces.AgentCollectionRequesterInterface
	agentCollectionsManagerRequester identity_interfaces.AgentCollectionsManagerRequesterInterface
	agentMinter                      identity_interfaces.AgentMinterInterface
	walletCredentialsStorage         identity_interfaces.WalletCredentialsStorageInterface
	agentRepository                  repositories.AgentRepositoryInterface
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface
	registrationReconciler           identity_interfaces.RegistrationReconcilerInterface
}

func NewMintPublisher(
	agentCollectionRequester identity_interfaces.AgentCollectionRequesterInterface,
	agentCollectionsManagerRequester identity_interfaces.AgentCollectionsManagerRequesterInterface,
	agentMinter identity_interfaces.AgentMinterInterface,
	walletCredentialsStorage identity_interfaces.WalletCredentialsStorageInterface,
	agentRepository repositories.AgentRepositoryInterface,
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface,
	registrationReconciler identity_interfaces.RegistrationReconcilerInterface,
) *MintPublisher {
	return &MintPublisher{
		agentCollectionRequester:         agentCollectionRequester,
		agentCollectionsManagerRequester: agentCollectionsManagerRequester,
		agentMinter:                      agentMinter,
		walletCredentialsStorage:         walletCredentialsStorage,
		agentRepository:                  agentRepository,
		agentRegistrationStateRepository: agentRegistrationStateRepository,
		registrationReconciler:           registrationReconciler,
	}
}

func (s *MintPublisher) SelfMint(
	ctx context.Context,
	agent *models.Agent,
	state *models.AgentRegistrationState,
) (*identity_interfaces.MintResult, error) {
	cred, found, err := s.walletCredentialsStorage.GetForId(ctx, agent.Id.String())
	if err != nil {
		return nil, fmt.Errorf("load node wallet: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("node wallet not bootstrapped")
	}
	nodeWallet := common.HexToAddress(cred.Address)
	collection := common.HexToAddress(agent.CollectionAddress)

	canMintDirectly, err := s.canMintDirectly(agent.ChainId, collection, nodeWallet)
	if err != nil {
		return nil, err
	}

	params := identity_interfaces.MintParams{
		Collection: collection,
		To:         common.HexToAddress(*agent.Owner),
		Role:       agent.Role.String(),
		Name:       agent.Name,
		Wallets:    []common.Address{nodeWallet},
		Images:     identity_interfaces.MintKeyValue{}.FromMap(agent.Images),
		Attributes: identity_interfaces.MintKeyValue{}.FromMap(agent.Attributes),
		CloneOf:    agent.CloneOf,
	}

	if !canMintDirectly {
		return s.createMintProposal(ctx, agent, state, params)
	}
	return s.createMint(ctx, agent, state, collection, nodeWallet, params)
}

func (s *MintPublisher) createMint(
	ctx context.Context,
	agent *models.Agent,
	state *models.AgentRegistrationState,
	collection common.Address,
	nodeWallet common.Address,
	params identity_interfaces.MintParams,
) (*identity_interfaces.MintResult, error) {
	txHash, err := s.agentMinter.Mint(ctx, agent.ChainId, params)
	if err != nil {
		return nil, fmt.Errorf("mint: %w", err)
	}
	tokenId, err := s.agentCollectionRequester.WalletToTokenId(agent.ChainId, collection, nodeWallet)
	if err != nil {
		return nil, fmt.Errorf("read minted token: %w", err)
	}
	if tokenId == 0 {
		return nil, fmt.Errorf("mint succeeded but token not found on chain")
	}

	savedAgent, savedState, err := s.registrationReconciler.ReconcileMint(ctx, agent, state, collection, tokenId, &txHash)
	if err != nil {
		return nil, fmt.Errorf("reconcile mint: %w", err)
	}
	return &identity_interfaces.MintResult{Agent: savedAgent, State: savedState}, nil
}

func (s *MintPublisher) ResolveMintedToken(ctx context.Context, agent *models.Agent) (common.Address, uint64, error) {
	cred, found, err := s.walletCredentialsStorage.GetForId(ctx, agent.Id.String())
	if err != nil {
		return common.Address{}, 0, fmt.Errorf("load node wallet: %w", err)
	}
	if !found {
		return common.Address{}, 0, fmt.Errorf("node wallet not bootstrapped")
	}
	nodeWallet := common.HexToAddress(cred.Address)
	collection := common.HexToAddress(agent.CollectionAddress)

	tokenId, err := s.agentCollectionRequester.WalletToTokenId(agent.ChainId, collection, nodeWallet)
	if err != nil {
		return common.Address{}, 0, fmt.Errorf("read minted token: %w", err)
	}

	// Fallback: the draft's collection may be wrong/empty; scan the manager.
	if tokenId == 0 {
		foundCollection, foundTokenId, _, findErr := s.agentCollectionsManagerRequester.FindAgentByWallet(agent.ChainId, nodeWallet)
		if findErr != nil {
			return common.Address{}, 0, fmt.Errorf("scan collections: %w", findErr)
		}
		if foundTokenId > 0 {
			collection = foundCollection
			tokenId = foundTokenId
		}
	}
	return collection, tokenId, nil
}

// canMintDirectly is true when the collection has no moderators or the node
// wallet is one of them; otherwise the mint must go through a proposal.
func (s *MintPublisher) canMintDirectly(chainId string, collection, nodeWallet common.Address) (bool, error) {
	moderatorCount, err := s.agentCollectionRequester.ModeratorCount(chainId, collection)
	if err != nil {
		return false, fmt.Errorf("read moderator count: %w", err)
	}
	if moderatorCount == 0 {
		return true, nil
	}
	isModerator, err := s.agentCollectionRequester.IsModerator(chainId, collection, nodeWallet)
	if err != nil {
		return false, fmt.Errorf("read moderator: %w", err)
	}
	return isModerator, nil
}

func (s *MintPublisher) createMintProposal(
	ctx context.Context,
	agent *models.Agent,
	state *models.AgentRegistrationState,
	params identity_interfaces.MintParams,
) (*identity_interfaces.MintResult, error) {
	txHash, proposalId, err := s.agentMinter.CreateMintProposal(ctx, agent.ChainId, params)
	if err != nil {
		return nil, fmt.Errorf("create mint proposal: %w", err)
	}

	agent.ProposalId = &proposalId
	savedAgent, err := s.agentRepository.Upsert(ctx, agent)
	if err != nil {
		return nil, fmt.Errorf("save proposal: %w", err)
	}

	if state == nil {
		state = &models.AgentRegistrationState{EnsRegistryStatus: enums.TransactionStatusPending}
	}
	state.MintStatus = enums.TransactionStatusSubmitted
	state.MintTxHash = &txHash
	savedState, err := s.agentRegistrationStateRepository.Upsert(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("save state: %w", err)
	}

	return &identity_interfaces.MintResult{Agent: savedAgent, State: savedState, ProposalPending: true}, nil
}
