package interfaces

import "context"

type WalletCredentials struct {
	PrivateKeyHex string // 0x-prefixed hex of the secp256k1 private key
	Address       string // 0x-prefixed checksummed Ethereum address
}

type WalletCredentialsStorageInterface interface {
	// Get returns the stored wallet. found=false means none persisted yet.
	Get(ctx context.Context) (cred *WalletCredentials, found bool, err error)
	// Save upserts the wallet credentials.
	Save(ctx context.Context, cred *WalletCredentials) error
	// GetOrGenerate returns the stored wallet, or generates+persists a new one.
	GetOrGenerate(ctx context.Context) (*WalletCredentials, error)
}
