package services

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
)

// RegistrationReconciler writes the agent + state rows from confirmed on-chain facts.
type RegistrationReconciler struct {
	agentCollectionRequester identity_interfaces.AgentCollectionRequesterInterface

	agentRepository                  repositories.AgentRepositoryInterface
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface
}

func NewRegistrationReconciler(
	agentRepository repositories.AgentRepositoryInterface,
	agentCollectionRequester identity_interfaces.AgentCollectionRequesterInterface,
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface,
) *RegistrationReconciler {
	return &RegistrationReconciler{
		agentRepository:                  agentRepository,
		agentCollectionRequester:         agentCollectionRequester,
		agentRegistrationStateRepository: agentRegistrationStateRepository,
	}
}

func (s *RegistrationReconciler) ReconcileMint(
	ctx context.Context,
	agent *models.Agent,
	state *models.AgentRegistrationState,
	collection common.Address,
	tokenId uint64,
	mintTxHash *string,
) (*models.Agent, *models.AgentRegistrationState, error) {
	info, err := s.agentCollectionRequester.InformationOf(agent.ChainId, collection, tokenId)
	if err != nil {
		return nil, nil, err
	}
	owner, err := s.agentCollectionRequester.OwnerOf(agent.ChainId, collection, tokenId)
	if err != nil {
		return nil, nil, err
	}

	ownerHex := owner.Hex()
	agent.CollectionAddress = collection.Hex()
	agent.CollectionTokenId = &tokenId
	agent.Creator = info.Creator.Hex()
	agent.Owner = &ownerHex
	savedAgent, err := s.agentRepository.Upsert(ctx, agent)
	if err != nil {
		return nil, nil, err
	}

	if state == nil {
		state = &models.AgentRegistrationState{EnsRegistryStatus: enums.TransactionStatusPending}
	}
	state.AgentId = savedAgent.Id
	state.MarkMintConfirmed()
	if mintTxHash != nil {
		state.MintTxHash = mintTxHash
	}
	savedState, err := s.agentRegistrationStateRepository.Upsert(ctx, state)
	if err != nil {
		return nil, nil, err
	}
	return savedAgent, savedState, nil
}

func (s *RegistrationReconciler) ReconcileEnsRegistry(
	ctx context.Context,
	agentId uuid.UUID,
	state *models.AgentRegistrationState,
) (*models.AgentRegistrationState, error) {
	if state == nil {
		state = &models.AgentRegistrationState{MintStatus: enums.TransactionStatusPending}
	}
	state.AgentId = agentId
	state.MarkEnsConfirmed()
	return s.agentRegistrationStateRepository.Upsert(ctx, state)
}
