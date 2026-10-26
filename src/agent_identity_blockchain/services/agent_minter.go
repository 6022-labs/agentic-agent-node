package services

import (
	"context"
	"fmt"
	"strconv"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/abi"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/factory"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/helper"
	common_helper "github.com/6022-labs/agentic-agent-node/src/runtime_common/helper"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

// AgentMinter sends mint transactions with the node's Vault wallet, once the owner has funded it.
type AgentMinter struct {
	logger            *zap.Logger
	ethClientFactory  *factory.EthClientFactory
	transactorFactory *factory.TransactorFactory
}

func NewAgentMinter(
	logger *zap.Logger,
	ethClientFactory *factory.EthClientFactory,
	transactorFactory *factory.TransactorFactory,
) *AgentMinter {
	return &AgentMinter{
		logger:            logger,
		ethClientFactory:  ethClientFactory,
		transactorFactory: transactorFactory,
	}
}

func (m *AgentMinter) Mint(ctx context.Context, chainId string, params identity_interfaces.MintParams) (string, error) {
	client, err := m.ethClientFactory.For(chainId)
	if err != nil {
		return "", err
	}
	evmChainId, err := strconv.ParseUint(chainId, 10, 64)
	if err != nil {
		return "", err
	}
	opts, err := m.transactorFactory.Create(ctx, evmChainId)
	if err != nil {
		return "", err
	}
	contract, err := abi.NewAgentCollectionV1(params.Collection, client)
	if err != nil {
		return "", fmt.Errorf("bind AgentCollectionV1: %w", err)
	}

	tx, err := contract.Mint(
		opts,
		params.To,
		params.Name,
		helper.ToAgentAddresses(params.Addresses),
		helper.ToKeyValues(params.Images),
		helper.ToKeyValues(params.Attributes),
		helper.ToNullableUint256(params.CloneOf),
	)
	if err != nil {
		return "", fmt.Errorf("mint: %w", err)
	}
	if err := common_helper.WaitForSuccess(ctx, client, tx); err != nil {
		return tx.Hash().Hex(), fmt.Errorf("mint: %w", err)
	}

	m.logger.Info("agent minted via node wallet",
		zap.String("collection", params.Collection.Hex()),
		zap.String("name", params.Name),
		zap.String("tx", tx.Hash().Hex()),
	)
	return tx.Hash().Hex(), nil
}

func (m *AgentMinter) CreateMintProposal(
	ctx context.Context,
	chainId string,
	params identity_interfaces.MintParams,
) (string, uint64, error) {
	client, err := m.ethClientFactory.For(chainId)
	if err != nil {
		return "", 0, err
	}
	evmChainId, err := strconv.ParseUint(chainId, 10, 64)
	if err != nil {
		return "", 0, err
	}
	opts, err := m.transactorFactory.Create(ctx, evmChainId)
	if err != nil {
		return "", 0, err
	}
	contract, err := abi.NewAgentCollectionV1(params.Collection, client)
	if err != nil {
		return "", 0, fmt.Errorf("bind AgentCollectionV1: %w", err)
	}

	tx, err := contract.CreateMintProposal(
		opts,
		params.To,
		params.Name,
		helper.ToAgentAddresses(params.Addresses),
		helper.ToKeyValues(params.Images),
		helper.ToKeyValues(params.Attributes),
		helper.ToNullableUint256(params.CloneOf),
	)
	if err != nil {
		return "", 0, fmt.Errorf("createMintProposal: %w", err)
	}
	receipt, err := common_helper.WaitForReceipt(ctx, client, tx)
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
