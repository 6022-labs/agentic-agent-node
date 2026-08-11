package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
)

// WalletCredentials is a keypair for one chain family: private key hex (the chain's own encoding) and derived address.
type WalletCredentials struct {
	PrivateKeyHex string
	Address       string
}

// WalletCredentialsStorageInterface stores the agent's wallet credentials per
// AddressType in Vault (path wallet/<type>), generating a missing one on demand.
type WalletCredentialsStorageInterface interface {
	// GetForId returns the stored credentials. found=false means none persisted yet.
	GetForId(
		ctx context.Context,
		agentId string,
		addressType enums.AddressType,
	) (cred *WalletCredentials, found bool, err error)
	// SaveForId upserts the credentials.
	SaveForId(ctx context.Context, agentId string, addressType enums.AddressType, cred *WalletCredentials) error
	// GetOrGenerateForId returns the stored credentials, or generates+persists new ones.
	GetOrGenerateForId(ctx context.Context, agentId string, addressType enums.AddressType) (*WalletCredentials, error)
}
