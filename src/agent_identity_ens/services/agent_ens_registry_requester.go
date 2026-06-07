package services

import (
	"context"
	"fmt"
	"math/big"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/factory"
	blockchainhelper "github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/helper"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_ens/abi"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_ens/helper"
	ens_settings "github.com/6022-labs/agentic-agent-node/src/agent_identity_ens/settings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// AgentEnsRegistryRequester is a thin translator over the on-chain
// AgentEnsRegistry: it resolves the agent's subnode and reads/writes records.
// It holds no domain state and touches no repositories — orchestration (mint
// guard, init vs refresh) lives in the EnsPublisher domain service.
type AgentEnsRegistryRequester struct {
	logger                   *zap.Logger
	client                   *ethclient.Client
	blockchainSettings       *ens_settings.EnsBlockchainSettings
	registrySettings         *ens_settings.AgentEnsRegistriesSettings
	transactorFactory        *factory.TransactorFactory
	agentCollectionRequester identity_interfaces.AgentCollectionRequesterInterface
}

func NewAgentEnsRegistryRequester(
	logger *zap.Logger,
	client *ethclient.Client,
	blockchainSettings *ens_settings.EnsBlockchainSettings,
	registrySettings *ens_settings.AgentEnsRegistriesSettings,
	transactorFactory *factory.TransactorFactory,
	agentCollectionRequester identity_interfaces.AgentCollectionRequesterInterface,
) *AgentEnsRegistryRequester {
	return &AgentEnsRegistryRequester{
		logger:                   logger,
		client:                   client,
		blockchainSettings:       blockchainSettings,
		registrySettings:         registrySettings,
		transactorFactory:        transactorFactory,
		agentCollectionRequester: agentCollectionRequester,
	}
}

func (r *AgentEnsRegistryRequester) Provisioned(ctx context.Context, collection common.Address, agentName string) (bool, error) {
	registry, err := r.registry()
	if err != nil {
		return false, err
	}
	node, err := r.resolveAgentNode(ctx, registry, collection, agentName)
	if err != nil {
		return false, err
	}
	return registry.Provisioned(&bind.CallOpts{Context: ctx}, node)
}

func (r *AgentEnsRegistryRequester) Text(ctx context.Context, collection common.Address, agentName, key string) (string, error) {
	registry, err := r.registry()
	if err != nil {
		return "", err
	}
	node, err := r.resolveAgentNode(ctx, registry, collection, agentName)
	if err != nil {
		return "", err
	}
	return registry.Text(&bind.CallOpts{Context: ctx}, node, key)
}

func (r *AgentEnsRegistryRequester) InitRecords(ctx context.Context, collection common.Address, tokenId *big.Int, records []identity_interfaces.EnsRecord) error {
	registry, err := r.registry()
	if err != nil {
		return err
	}
	opts, err := r.transactorFactory.Create(ctx, r.blockchainSettings.ChainId)
	if err != nil {
		return err
	}

	abiRecords := make([]abi.KeyValue, 0, len(records))
	for _, record := range records {
		abiRecords = append(abiRecords, abi.KeyValue{Key: record.Key, Value: record.Value})
	}

	tx, err := registry.InitAgentTexts(opts, collection, tokenId, abiRecords)
	if err != nil {
		return fmt.Errorf("initAgentTexts: %w", err)
	}
	if err := blockchainhelper.WaitForSuccess(ctx, r.client, tx); err != nil {
		return fmt.Errorf("initAgentTexts: %w", err)
	}

	r.logger.Info("agent ENS subnode provisioned",
		zap.String("collection", collection.Hex()),
		zap.String("tx", tx.Hash().Hex()),
	)
	return nil
}

func (r *AgentEnsRegistryRequester) SetRecord(ctx context.Context, collection common.Address, tokenId *big.Int, key, value string) error {
	registry, err := r.registry()
	if err != nil {
		return err
	}
	opts, err := r.transactorFactory.Create(ctx, r.blockchainSettings.ChainId)
	if err != nil {
		return err
	}

	tx, err := registry.SetAgentText(opts, collection, tokenId, key, value)
	if err != nil {
		return fmt.Errorf("setAgentText %s: %w", key, err)
	}
	if err := blockchainhelper.WaitForSuccess(ctx, r.client, tx); err != nil {
		return fmt.Errorf("setAgentText %s: %w", key, err)
	}

	r.logger.Info("agent ENS text record set",
		zap.String("collection", collection.Hex()),
		zap.String("key", key),
		zap.String("tx", tx.Hash().Hex()),
	)
	return nil
}

func (r *AgentEnsRegistryRequester) registry() (*abi.AgentEnsRegistry, error) {
	registry, err := abi.NewAgentEnsRegistry(r.registrySettings.RegistryAddress, r.client)
	if err != nil {
		return nil, fmt.Errorf("bind AgentEnsRegistry: %w", err)
	}
	return registry, nil
}

// resolveAgentNode reads parentNode + collection name on chain, then computes
// the agent's ENS namehash.
func (r *AgentEnsRegistryRequester) resolveAgentNode(
	ctx context.Context,
	registry *abi.AgentEnsRegistry,
	collection common.Address,
	agentLabel string,
) ([32]byte, error) {
	parentNode, err := registry.ParentNode(&bind.CallOpts{Context: ctx})
	if err != nil {
		return [32]byte{}, fmt.Errorf("read parentNode: %w", err)
	}

	collectionName, err := r.agentCollectionRequester.CollectionName(collection)
	if err != nil {
		return [32]byte{}, fmt.Errorf("read collection name: %w", err)
	}

	return helper.ResolveAgentNode(parentNode, r.blockchainSettings.ChainId, collectionName, agentLabel), nil
}
