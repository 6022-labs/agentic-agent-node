package services

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	blockchain_abi "github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/abi"
	blockchain_settings "github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_ens/abi"
	ens_settings "github.com/6022-labs/agentic-agent-node/src/agent_identity_ens/settings"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/models"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

const (
	urlKey = "url"
)

// AgentEnsRegistryRequester publishes the agent's runtime URL to its ENS
// subnode via the on-chain AgentEnsRegistry, at boot once wallet + identity are
// ready. First boot seeds the keys via initAgentTexts; later boots refresh
// individual records via setAgentText (no-op if unchanged).
type AgentEnsRegistryRequester struct {
	logger             *zap.Logger
	client             *ethclient.Client
	blockchainSettings *blockchain_settings.BlockchainSettings
	registrySettings   *ens_settings.AgentEnsRegistriesSettings
	ensSettings        *ens_settings.AgentIdentityEnsSettings
	walletStorage      identity_interfaces.WalletCredentialsStorageInterface
	identityResolver   common_interfaces.AgentIdentityResolverInterface
}

func NewAgentEnsRegistryRequester(
	logger *zap.Logger,
	client *ethclient.Client,
	blockchainSettings *blockchain_settings.BlockchainSettings,
	registrySettings *ens_settings.AgentEnsRegistriesSettings,
	ensSettings *ens_settings.AgentIdentityEnsSettings,
	walletStorage identity_interfaces.WalletCredentialsStorageInterface,
	identityResolver common_interfaces.AgentIdentityResolverInterface,
) *AgentEnsRegistryRequester {
	return &AgentEnsRegistryRequester{
		logger:             logger,
		client:             client,
		blockchainSettings: blockchainSettings,
		registrySettings:   registrySettings,
		ensSettings:        ensSettings,
		walletStorage:      walletStorage,
		identityResolver:   identityResolver,
	}
}

// PublishRuntimeRecords writes the `url` text record for the agent's NFT.
// Requires the NFT to already exist on chain.
func (r *AgentEnsRegistryRequester) PublishRuntimeRecords(ctx context.Context) error {
	identity, err := r.identityResolver.Resolve(ctx)
	if err != nil {
		return fmt.Errorf("resolve agent identity: %w", err)
	}

	cred, found, err := r.walletStorage.Get(ctx)
	if err != nil {
		return fmt.Errorf("load wallet credentials: %w", err)
	}
	if !found {
		return fmt.Errorf("wallet credentials not present; bootstrap must complete before ENS publish")
	}

	opts, err := r.newTransactor(cred)
	if err != nil {
		return err
	}
	opts.Context = ctx

	registry, err := abi.NewAgentEnsRegistry(r.registrySettings.RegistryAddress, r.client)
	if err != nil {
		return fmt.Errorf("bind AgentEnsRegistry: %w", err)
	}

	collectionAddr := common.HexToAddress(identity.CollectionAddress)
	tokenId := new(big.Int).SetUint64(identity.TokenId)

	node, err := r.resolveAgentNode(ctx, registry, collectionAddr, identity.Name)
	if err != nil {
		return fmt.Errorf("resolve agent node: %w", err)
	}

	provisioned, err := registry.Provisioned(&bind.CallOpts{Context: ctx}, node)
	if err != nil {
		return fmt.Errorf("read provisioned: %w", err)
	}

	if !provisioned {
		return r.initAgentTexts(ctx, registry, opts, collectionAddr, tokenId, identity)
	}
	return r.refreshAgentTexts(ctx, registry, opts, collectionAddr, tokenId, node, identity)
}

func (r *AgentEnsRegistryRequester) initAgentTexts(
	ctx context.Context,
	registry *abi.AgentEnsRegistry,
	opts *bind.TransactOpts,
	collection common.Address,
	tokenId *big.Int,
	identity *models.AgentIdentity,
) error {
	records := []abi.KeyValue{
		{Key: urlKey, Value: r.ensSettings.RuntimeUrl},
	}

	tx, err := registry.InitAgentTexts(opts, collection, tokenId, records)
	if err != nil {
		return fmt.Errorf("initAgentTexts: %w", err)
	}
	if err := r.waitForSuccess(ctx, tx); err != nil {
		return fmt.Errorf("initAgentTexts: %w", err)
	}

	r.logger.Info("agent ENS subnode provisioned",
		zap.String("collection", identity.CollectionAddress),
		zap.Uint64("tokenId", identity.TokenId),
		zap.String("name", identity.Name),
		zap.String("runtimeUrl", r.ensSettings.RuntimeUrl),
		zap.String("tx", tx.Hash().Hex()),
	)
	return nil
}

func (r *AgentEnsRegistryRequester) refreshAgentTexts(
	ctx context.Context,
	registry *abi.AgentEnsRegistry,
	opts *bind.TransactOpts,
	collection common.Address,
	tokenId *big.Int,
	node [32]byte,
	identity *models.AgentIdentity,
) error {
	records := []struct {
		key   string
		value string
	}{
		{urlKey, r.ensSettings.RuntimeUrl},
	}

	callOpts := &bind.CallOpts{Context: ctx}
	for _, rec := range records {
		current, err := registry.Text(callOpts, node, rec.key)
		if err != nil {
			return fmt.Errorf("read text %s: %w", rec.key, err)
		}
		if current == rec.value {
			r.logger.Debug("agent ENS text record already up to date",
				zap.String("collection", identity.CollectionAddress),
				zap.Uint64("tokenId", identity.TokenId),
				zap.String("name", identity.Name),
				zap.String("key", rec.key),
			)
			continue
		}

		tx, err := registry.SetAgentText(opts, collection, tokenId, rec.key, rec.value)
		if err != nil {
			return fmt.Errorf("setAgentText %s: %w", rec.key, err)
		}
		if err := r.waitForSuccess(ctx, tx); err != nil {
			return fmt.Errorf("setAgentText %s: %w", rec.key, err)
		}
		r.logger.Info("agent ENS text record refreshed",
			zap.String("collection", identity.CollectionAddress),
			zap.Uint64("tokenId", identity.TokenId),
			zap.String("name", identity.Name),
			zap.String("key", rec.key),
			zap.String("previous", current),
			zap.String("value", rec.value),
			zap.String("tx", tx.Hash().Hex()),
		)
	}
	return nil
}

// resolveAgentNode mirrors AgentEnsRegistry._resolveAgentNode: walks the
// namehash parentNode → chainId label → collection name → agent label, where
// each step is keccak256(prev || keccak256(labelBytes)).
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

	collectionContract, err := blockchain_abi.NewAgentCollectionV1(collection, r.client)
	if err != nil {
		return [32]byte{}, fmt.Errorf("bind AgentCollectionV1: %w", err)
	}
	collectionName, err := collectionContract.Name(&bind.CallOpts{Context: ctx})
	if err != nil {
		return [32]byte{}, fmt.Errorf("read collection name: %w", err)
	}

	chainIdLabel := crypto.Keccak256Hash([]byte(strconv.FormatUint(r.blockchainSettings.Id, 10)))
	chainNode := crypto.Keccak256Hash(parentNode[:], chainIdLabel[:])
	collectionLabel := crypto.Keccak256Hash([]byte(collectionName))
	collectionNode := crypto.Keccak256Hash(chainNode[:], collectionLabel[:])
	agentLabelHash := crypto.Keccak256Hash([]byte(agentLabel))
	return crypto.Keccak256Hash(collectionNode[:], agentLabelHash[:]), nil
}

func (r *AgentEnsRegistryRequester) waitForSuccess(ctx context.Context, tx *types.Transaction) error {
	receipt, err := bind.WaitMined(ctx, r.client, tx)
	if err != nil {
		return fmt.Errorf("wait for tx %s: %w", tx.Hash().Hex(), err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("tx %s reverted (status=%d)", tx.Hash().Hex(), receipt.Status)
	}
	return nil
}

func (r *AgentEnsRegistryRequester) newTransactor(cred *identity_interfaces.WalletCredentials) (*bind.TransactOpts, error) {
	key, err := privateKeyFromCredentials(cred)
	if err != nil {
		return nil, err
	}
	return bind.NewKeyedTransactorWithChainID(key, new(big.Int).SetUint64(r.blockchainSettings.Id))
}

func privateKeyFromCredentials(cred *identity_interfaces.WalletCredentials) (*ecdsa.PrivateKey, error) {
	hex := cred.PrivateKeyHex
	if len(hex) >= 2 && (hex[:2] == "0x" || hex[:2] == "0X") {
		hex = hex[2:]
	}
	hex = strings.TrimSpace(hex)
	key, err := crypto.HexToECDSA(hex)
	if err != nil {
		return nil, fmt.Errorf("invalid wallet private key: %w", err)
	}
	return key, nil
}
