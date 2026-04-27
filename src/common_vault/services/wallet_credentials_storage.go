package services

import (
	"context"
	"errors"
	"fmt"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	vault_interfaces "github.com/6022-labs/agentic-agent-node/src/common_vault/services/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	walletSecretKey         = "agent/wallet"
	walletFieldPrivateKey   = "privateKey"
	walletFieldAddress      = "address"
)

type WalletCredentialsStorage struct {
	storage vault_interfaces.KvV1SecretStorageInterface
}

func NewWalletCredentialsStorage(storage vault_interfaces.KvV1SecretStorageInterface) *WalletCredentialsStorage {
	return &WalletCredentialsStorage{storage: storage}
}

func (w *WalletCredentialsStorage) Get(ctx context.Context) (*identity_interfaces.WalletCredentials, bool, error) {
	data, err := w.storage.Read(ctx, walletSecretKey)
	if errors.Is(err, ErrSecretNotFound) {
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

func (w *WalletCredentialsStorage) Save(ctx context.Context, cred *identity_interfaces.WalletCredentials) error {
	if cred == nil || len(cred.PrivateKeyHex) == 0 || len(cred.Address) == 0 {
		return errors.New("wallet credentials must have both privateKey and address")
	}
	return w.storage.Write(ctx, walletSecretKey, map[string]string{
		walletFieldPrivateKey: cred.PrivateKeyHex,
		walletFieldAddress:    cred.Address,
	})
}

func (w *WalletCredentialsStorage) GetOrGenerate(ctx context.Context) (*identity_interfaces.WalletCredentials, error) {
	cred, found, err := w.Get(ctx)
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
	if err := w.Save(ctx, cred); err != nil {
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
