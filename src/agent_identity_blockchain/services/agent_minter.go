package services

import (
	"context"
	"fmt"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/abi"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/factory"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/helper"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// AgentMinter sends mint transactions with the node's bootstrap wallet (from
// Vault). Used by the self-mint flow once the owner has funded that wallet.
type AgentMinter struct {
	logger             *zap.Logger
	client             *ethclient.Client
	blockchainSettings *settings.BlockchainSettings
	transactorFactory  *factory.TransactorFactory
}

func NewAgentMinter(
	logger *zap.Logger,
	client *ethclient.Client,
	blockchainSettings *settings.BlockchainSettings,
	transactorFactory *factory.TransactorFactory,
) *AgentMinter {
	return &AgentMinter{
		logger:             logger,
		client:             client,
		blockchainSettings: blockchainSettings,
		transactorFactory:  transactorFactory,
	}
}

func (m *AgentMinter) Mint(ctx context.Context, params identity_interfaces.MintParams) (string, error) {
	opts, err := m.transactorFactory.Create(ctx, m.blockchainSettings.Id)
	if err != nil {
		return "", err
	}
	contract, err := abi.NewAgentCollectionV1(params.Collection, m.client)
	if err != nil {
		return "", fmt.Errorf("bind AgentCollectionV1: %w", err)
	}

	tx, err := contract.Mint(opts, params.To, params.Role, params.Name, params.Wallets, helper.ToKeyValues(params.Images), helper.ToKeyValues(params.Attributes), helper.ToNullableUint256(params.CloneOf))
	if err != nil {
		return "", fmt.Errorf("mint: %w", err)
	}
	if err := helper.WaitForSuccess(ctx, m.client, tx); err != nil {
		return tx.Hash().Hex(), fmt.Errorf("mint: %w", err)
	}

	m.logger.Info("agent minted via node wallet",
		zap.String("collection", params.Collection.Hex()),
		zap.String("name", params.Name),
		zap.String("tx", tx.Hash().Hex()),
	)
	return tx.Hash().Hex(), nil
}

func (m *AgentMinter) CreateMintProposal(ctx context.Context, params identity_interfaces.MintParams) (string, uint64, error) {
	opts, err := m.transactorFactory.Create(ctx, m.blockchainSettings.Id)
	if err != nil {
		return "", 0, err
	}
	contract, err := abi.NewAgentCollectionV1(params.Collection, m.client)
	if err != nil {
		return "", 0, fmt.Errorf("bind AgentCollectionV1: %w", err)
	}

	tx, err := contract.CreateMintProposal(opts, params.To, params.Role, params.Name, params.Wallets, helper.ToKeyValues(params.Images), helper.ToKeyValues(params.Attributes), helper.ToNullableUint256(params.CloneOf))
	if err != nil {
		return "", 0, fmt.Errorf("createMintProposal: %w", err)
	}
	receipt, err := helper.WaitForReceipt(ctx, m.client, tx)
	if err != nil {
		return tx.Hash().Hex(), 0, fmt.Errorf("createMintProposal: %w", err)
	}

	proposalId, err := m.parseProposalId(contract, receipt)
	if err != nil {
		return tx.Hash().Hex(), 0, fmt.Errorf("createMintProposal: %w", err)
	}

	m.logger.Info("agent mint proposal created via node wallet",
		zap.String("collection", params.Collection.Hex()),
		zap.String("name", params.Name),
		zap.Uint64("proposalId", proposalId),
		zap.String("tx", tx.Hash().Hex()),
	)
	return tx.Hash().Hex(), proposalId, nil
}

func (m *AgentMinter) parseProposalId(contract *abi.AgentCollectionV1, receipt *types.Receipt) (uint64, error) {
	for _, log := range receipt.Logs {
		event, err := contract.ParseMintProposalCreated(*log)
		if err != nil {
			continue // not this event
		}
		if event.ProposalId != nil {
			return event.ProposalId.Uint64(), nil
		}
	}
	return 0, fmt.Errorf("MintProposalCreated event not found in receipt")
}
