package interfaces

import "context"

type WalletCredentials struct {
	PrivateKeyHex string // 0x-prefixed hex of the secp256k1 private key
	Address       string // 0x-prefixed checksummed Ethereum address
}

type WalletCredentialsStorageInterface interface {
	// GetForId returns the agent's stored wallet. found=false means none persisted yet.
	GetForId(ctx context.Context, agentId string) (cred *WalletCredentials, found bool, err error)
	// SaveForId upserts the agent's wallet credentials.
	SaveForId(ctx context.Context, agentId string, cred *WalletCredentials) error
	// GetOrGenerateForId returns the agent's stored wallet, or generates+persists a new one.
	GetOrGenerateForId(ctx context.Context, agentId string) (*WalletCredentials, error)
}
