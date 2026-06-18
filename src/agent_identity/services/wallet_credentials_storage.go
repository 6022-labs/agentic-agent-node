package services

import (
	"context"
	"errors"
	"fmt"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	walletSecretKey       = "wallet"
	walletFieldPrivateKey = "privateKey"
	walletFieldAddress    = "address"
)

type WalletCredentialsStorage struct {
	storage               runtime_common_interfaces.SecretStorageInterface
	agentScopedKeyBuilder identity_interfaces.AgentScopedKeyBuilderInterface
}

func NewWalletCredentialsStorage(
	storage runtime_common_interfaces.SecretStorageInterface,
	agentScopedKeyBuilder identity_interfaces.AgentScopedKeyBuilderInterface,
) *WalletCredentialsStorage {
	return &WalletCredentialsStorage{storage: storage, agentScopedKeyBuilder: agentScopedKeyBuilder}
}

func (w *WalletCredentialsStorage) GetForId(ctx context.Context, agentId string) (*identity_interfaces.WalletCredentials, bool, error) {
	key := w.agentScopedKeyBuilder.ScopedKeyFor(agentId, walletSecretKey)
	data, err := w.storage.Read(ctx, key)
	if errors.Is(err, runtime_common_interfaces.ErrSecretNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	pk, ok := data[walletFieldPrivateKey]
	if !ok || len(pk) == 0 {
		return nil, false, fmt.Errorf("wallet secret missing %q field", walletFieldPrivateKey)
	}
	addr := data[walletFieldAddress]
	if len(addr) == 0 {
		// Backfill from private key if absent (defensive).
		derived, derr := deriveAddress(pk)
		if derr != nil {
			return nil, false, fmt.Errorf("wallet secret has no address and derivation failed: %w", derr)
		}
		addr = derived
	}

	return &identity_interfaces.WalletCredentials{PrivateKeyHex: pk, Address: addr}, true, nil
}

func (w *WalletCredentialsStorage) SaveForId(ctx context.Context, agentId string, cred *identity_interfaces.WalletCredentials) error {
	if cred == nil || len(cred.PrivateKeyHex) == 0 || len(cred.Address) == 0 {
		return errors.New("wallet credentials must have both privateKey and address")
	}
	key := w.agentScopedKeyBuilder.ScopedKeyFor(agentId, walletSecretKey)
	return w.storage.Write(ctx, key, map[string]string{
		walletFieldPrivateKey: cred.PrivateKeyHex,
		walletFieldAddress:    cred.Address,
	})
}

func (w *WalletCredentialsStorage) GetOrGenerateForId(ctx context.Context, agentId string) (*identity_interfaces.WalletCredentials, error) {
	cred, found, err := w.GetForId(ctx, agentId)
	if err != nil {
		return nil, err
	}
	if found {
		return cred, nil
	}

	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate wallet key: %w", err)
	}

	pkBytes := crypto.FromECDSA(key)
	pkHex := "0x" + common.Bytes2Hex(pkBytes)
	addr := crypto.PubkeyToAddress(key.PublicKey).Hex()

	cred = &identity_interfaces.WalletCredentials{PrivateKeyHex: pkHex, Address: addr}
	if err := w.SaveForId(ctx, agentId, cred); err != nil {
		return nil, fmt.Errorf("failed to persist generated wallet: %w", err)
	}
	return cred, nil
}

func deriveAddress(privateKeyHex string) (string, error) {
	pk := privateKeyHex
	if len(pk) >= 2 && (pk[:2] == "0x" || pk[:2] == "0X") {
		pk = pk[2:]
	}
	key, err := crypto.HexToECDSA(pk)
	if err != nil {
		return "", err
	}
	return crypto.PubkeyToAddress(key.PublicKey).Hex(), nil
}
