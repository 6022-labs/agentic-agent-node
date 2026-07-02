package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	"go.uber.org/dig"
)

const (
	walletSecretKeyPrefix = "wallet"
	walletFieldPrivateKey = "privateKey"
	walletFieldAddress    = "address"
)

// WalletCredentialsStorage stores one keypair per AddressType in Vault under agents/{id}/wallet/<type>.
type WalletCredentialsStorage struct {
	storage               runtime_common_interfaces.SecretStorageInterface
	agentScopedKeyBuilder identity_interfaces.AgentScopedKeyBuilderInterface
	generators            map[enums.AddressType]identity_interfaces.AddressKeyGeneratorInterface
}

type NewWalletCredentialsStorageParams struct {
	dig.In

	Storage               runtime_common_interfaces.SecretStorageInterface
	AgentScopedKeyBuilder identity_interfaces.AgentScopedKeyBuilderInterface
	Generators            []identity_interfaces.AddressKeyGeneratorInterface `group:"address_key_generators"`
}

func NewWalletCredentialsStorage(params NewWalletCredentialsStorageParams) *WalletCredentialsStorage {
	generators := map[enums.AddressType]identity_interfaces.AddressKeyGeneratorInterface{}
	for _, generator := range params.Generators {
		generators[generator.Type()] = generator
	}

	return &WalletCredentialsStorage{
		storage:               params.Storage,
		agentScopedKeyBuilder: params.AgentScopedKeyBuilder,
		generators:            generators,
	}
}

func (w *WalletCredentialsStorage) GetForId(ctx context.Context, agentId string, addressType enums.AddressType) (*identity_interfaces.WalletCredentials, bool, error) {
	data, err := w.storage.Read(ctx, w.vaultKey(agentId, addressType))
	if errors.Is(err, runtime_common_interfaces.ErrSecretNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	pk := data[walletFieldPrivateKey]
	if len(pk) == 0 {
		return nil, false, fmt.Errorf("%s wallet secret missing %q field", addressType, walletFieldPrivateKey)
	}
	addr := data[walletFieldAddress]
	if len(addr) == 0 {
		return nil, false, fmt.Errorf("%s wallet secret missing %q field", addressType, walletFieldAddress)
	}

	return &identity_interfaces.WalletCredentials{PrivateKeyHex: pk, Address: addr}, true, nil
}

func (w *WalletCredentialsStorage) SaveForId(ctx context.Context, agentId string, addressType enums.AddressType, cred *identity_interfaces.WalletCredentials) error {
	if cred == nil || len(cred.PrivateKeyHex) == 0 || len(cred.Address) == 0 {
		return errors.New("wallet credentials must have both privateKey and address")
	}
	return w.storage.Write(ctx, w.vaultKey(agentId, addressType), map[string]string{
		walletFieldPrivateKey: cred.PrivateKeyHex,
		walletFieldAddress:    cred.Address,
	})
}

func (w *WalletCredentialsStorage) GetOrGenerateForId(ctx context.Context, agentId string, addressType enums.AddressType) (*identity_interfaces.WalletCredentials, error) {
	cred, found, err := w.GetForId(ctx, agentId, addressType)
	if err != nil {
		return nil, err
	}
	if found {
		return cred, nil
	}

	generator, ok := w.generators[addressType]
	if !ok {
		return nil, fmt.Errorf("no key generator for address type %q", addressType)
	}
	cred, err = generator.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate %s key: %w", addressType, err)
	}
	if err := w.SaveForId(ctx, agentId, addressType, cred); err != nil {
		return nil, fmt.Errorf("failed to persist generated %s wallet: %w", addressType, err)
	}
	return cred, nil
}

// vaultKey is agents/{id}/wallet/<type> — the AddressType is the path segment.
func (w *WalletCredentialsStorage) vaultKey(agentId string, addressType enums.AddressType) string {
	return w.agentScopedKeyBuilder.ScopedKeyFor(agentId, walletSecretKeyPrefix+"/"+addressType.String())
}
