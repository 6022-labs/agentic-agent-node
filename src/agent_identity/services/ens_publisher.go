package services

import (
	"context"
	"fmt"
	"math/big"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// EnsPublisher orchestrates ENS publishing for the agent: it loads the agent +
// registration state, guards on mint, and decides whether to seed or refresh
// records, delegating the chain reads/writes to EnsRegistrarInterface.
type EnsPublisher struct {
	logger                           *zap.Logger
	ensRegistrar                     identity_interfaces.EnsRegistrarInterface
	agentRepository                  repositories.AgentRepositoryInterface
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface
	runtimeRecordsProvider           identity_interfaces.RuntimeRecordsProviderInterface
}

func NewEnsPublisher(
	logger *zap.Logger,
	ensRegistrar identity_interfaces.EnsRegistrarInterface,
	agentRepository repositories.AgentRepositoryInterface,
	runtimeRecordsProvider identity_interfaces.RuntimeRecordsProviderInterface,
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface,
) *EnsPublisher {
	return &EnsPublisher{
		logger:                           logger,
		ensRegistrar:                     ensRegistrar,
		agentRepository:                  agentRepository,
		agentRegistrationStateRepository: agentRegistrationStateRepository,
		runtimeRecordsProvider:           runtimeRecordsProvider,
	}
}

func (s *EnsPublisher) PublishRuntimeRecords(ctx context.Context) error {
	agent, err := s.agentRepository.Get(ctx)
	if err != nil {
		return fmt.Errorf("load agent: %w", err)
	}
	if agent == nil || agent.CollectionTokenId == nil {
		return common_interfaces.ErrIdentityNotFound
	}

	state, err := s.agentRegistrationStateRepository.GetByAgentId(ctx, agent.Id)
	if err != nil {
		return fmt.Errorf("load registration state: %w", err)
	}
	if !state.IsMinted() {
		return common_interfaces.ErrIdentityNotFound
	}

	collection := common.HexToAddress(agent.CollectionAddress)
	tokenId := new(big.Int).SetUint64(*agent.CollectionTokenId)

	provisioned, err := s.ensRegistrar.Provisioned(ctx, agent.ChainId, collection, agent.Name)
	if err != nil {
		return err
	}

	if !provisioned {
		return s.ensRegistrar.InitRecords(ctx, agent.ChainId, collection, tokenId, s.runtimeRecordsProvider.RuntimeRecords())
	}

	for _, record := range s.runtimeRecordsProvider.RuntimeRecords() {
		current, err := s.ensRegistrar.Text(ctx, agent.ChainId, collection, agent.Name, record.Key)
		if err != nil {
			return err
		}
		if current == record.Value {
			continue
		}
		if err := s.ensRegistrar.SetRecord(ctx, agent.ChainId, collection, tokenId, record.Key, record.Value); err != nil {
			return err
		}
		s.logger.Info("agent ENS text record refreshed",
			zap.String("name", agent.Name),
			zap.String("key", record.Key),
			zap.String("previous", current),
			zap.String("value", record.Value),
		)
	}
	return nil
}

func (s *EnsPublisher) IsProvisioned(ctx context.Context) (bool, error) {
	agent, err := s.agentRepository.Get(ctx)
	if err != nil {
		return false, fmt.Errorf("load agent: %w", err)
	}
	if agent == nil || agent.CollectionTokenId == nil {
		return false, common_interfaces.ErrIdentityNotFound
	}
	return s.ensRegistrar.Provisioned(ctx, agent.ChainId, common.HexToAddress(agent.CollectionAddress), agent.Name)
}
