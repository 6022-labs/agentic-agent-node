package services

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	abi "github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/abi_ens"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/factory"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/helper"
	ens_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// AgentEnsRegistryRequester is a thin translator over the on-chain
// AgentEnsRegistry: it resolves the agent's subnode and reads/writes records.
// It holds no domain state and touches no repositories — orchestration (mint
// guard, init vs refresh) lives in the EnsPublisher domain service.
type AgentEnsRegistryRequester struct {
	logger                   *zap.Logger
	ethClientFactory         *factory.EthClientFactory
	registrySettings         *ens_settings.AgentEnsRegistriesSettings
	transactorFactory        *factory.TransactorFactory
	agentCollectionRequester identity_interfaces.AgentCollectionRequesterInterface
}

func NewAgentEnsRegistryRequester(
	logger *zap.Logger,
	ethClientFactory *factory.EthClientFactory,
	registrySettings *ens_settings.AgentEnsRegistriesSettings,
	transactorFactory *factory.TransactorFactory,
	agentCollectionRequester identity_interfaces.AgentCollectionRequesterInterface,
) *AgentEnsRegistryRequester {
	return &AgentEnsRegistryRequester{
		logger:                   logger,
		ethClientFactory:         ethClientFactory,
		registrySettings:         registrySettings,
		transactorFactory:        transactorFactory,
		agentCollectionRequester: agentCollectionRequester,
	}
}

func (r *AgentEnsRegistryRequester) Provisioned(ctx context.Context, chainId string, collection common.Address, agentName string) (bool, error) {
	registry, err := r.registry(chainId)
	if err != nil {
		return false, err
	}
	node, err := r.resolveAgentNode(ctx, chainId, registry, collection, agentName)
	if err != nil {
		return false, err
	}
	return registry.Provisioned(&bind.CallOpts{Context: ctx}, node)
}

func (r *AgentEnsRegistryRequester) Text(ctx context.Context, chainId string, collection common.Address, agentName, key string) (string, error) {
	registry, err := r.registry(chainId)
	if err != nil {
		return "", err
	}
	node, err := r.resolveAgentNode(ctx, chainId, registry, collection, agentName)
	if err != nil {
		return "", err
	}
	return registry.Text(&bind.CallOpts{Context: ctx}, node, key)
}

func (r *AgentEnsRegistryRequester) InitRecords(ctx context.Context, chainId string, collection common.Address, tokenId *big.Int, records []identity_interfaces.EnsRecord) error {
	client, err := r.ethClientFactory.For(chainId)
	if err != nil {
		return err
	}
	registry, err := r.registry(chainId)
	if err != nil {
		return err
	}
	opts, err := r.transactor(ctx, chainId)
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
	if err := helper.WaitForSuccess(ctx, client, tx); err != nil {
		return fmt.Errorf("initAgentTexts: %w", err)
	}

	r.logger.Info("agent ENS subnode provisioned",
		zap.String("collection", collection.Hex()),
		zap.String("tx", tx.Hash().Hex()),
	)
	return nil
}

func (r *AgentEnsRegistryRequester) SetRecord(ctx context.Context, chainId string, collection common.Address, tokenId *big.Int, key, value string) error {
	client, err := r.ethClientFactory.For(chainId)
	if err != nil {
		return err
	}
	registry, err := r.registry(chainId)
	if err != nil {
		return err
	}
	opts, err := r.transactor(ctx, chainId)
	if err != nil {
		return err
	}

	tx, err := registry.SetAgentText(opts, collection, tokenId, key, value)
	if err != nil {
		return fmt.Errorf("setAgentText %s: %w", key, err)
	}
	if err := helper.WaitForSuccess(ctx, client, tx); err != nil {
		return fmt.Errorf("setAgentText %s: %w", key, err)
	}

	r.logger.Info("agent ENS text record set",
		zap.String("collection", collection.Hex()),
		zap.String("key", key),
		zap.String("tx", tx.Hash().Hex()),
	)
	return nil
}

func (r *AgentEnsRegistryRequester) registry(chainId string) (*abi.AgentEnsRegistry, error) {
	address, ok := r.registrySettings.AddressFor(chainId)
	if !ok {
		return nil, fmt.Errorf("no AgentEnsRegistry configured for chainId %s", chainId)
	}
	client, err := r.ethClientFactory.For(chainId)
	if err != nil {
		return nil, err
	}
	registry, err := abi.NewAgentEnsRegistry(address, client)
	if err != nil {
		return nil, fmt.Errorf("bind AgentEnsRegistry: %w", err)
	}
	return registry, nil
}

// transactor builds a signing transactor bound to the chain's EVM id.
func (r *AgentEnsRegistryRequester) transactor(ctx context.Context, chainId string) (*bind.TransactOpts, error) {
	evmChainId, err := strconv.ParseUint(chainId, 10, 64)
	if err != nil {
		return nil, err
	}
	return r.transactorFactory.Create(ctx, evmChainId)
}

// resolveAgentNode reads parentNode + collection name on chain, then computes
// the agent's ENS namehash.
func (r *AgentEnsRegistryRequester) resolveAgentNode(
	ctx context.Context,
	chainId string,
	registry *abi.AgentEnsRegistry,
	collection common.Address,
	agentLabel string,
) ([32]byte, error) {
	evmChainId, err := strconv.ParseUint(chainId, 10, 64)
	if err != nil {
		return [32]byte{}, err
	}
	parentNode, err := registry.ParentNode(&bind.CallOpts{Context: ctx})
	if err != nil {
		return [32]byte{}, fmt.Errorf("read parentNode: %w", err)
	}

	collectionName, err := r.agentCollectionRequester.CollectionName(chainId, collection)
	if err != nil {
		return [32]byte{}, fmt.Errorf("read collection name: %w", err)
	}

	return helper.ResolveAgentNode(parentNode, evmChainId, collectionName, agentLabel), nil
}
