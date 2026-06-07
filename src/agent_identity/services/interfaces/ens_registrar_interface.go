package interfaces

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// UrlRecordKey is the standard ENS text-record key for the agent's runtime URL.
// It is fixed by the contract, not configurable.
const UrlRecordKey = "url"

// EnsRecord is one text record key/value the agent publishes to its ENS subnode.
type EnsRecord struct {
	Key   string
	Value string
}

// EnsRegistrarInterface is a thin translator over the on-chain AgentEnsRegistry.
// It resolves the agent's subnode and reads/writes records; it holds no domain
// state and touches no repositories. Orchestration lives in EnsPublisher.
type EnsRegistrarInterface interface {
	// Provisioned reports whether the agent's subnode already has records.
	Provisioned(ctx context.Context, collection common.Address, agentName string) (bool, error)
	// Text reads a single text record off the agent's subnode.
	Text(ctx context.Context, collection common.Address, agentName, key string) (string, error)
	// InitRecords seeds the agent's subnode with its initial records.
	InitRecords(ctx context.Context, collection common.Address, tokenId *big.Int, records []EnsRecord) error
	// SetRecord overwrites a single record on the agent's subnode.
	SetRecord(ctx context.Context, collection common.Address, tokenId *big.Int, key, value string) error
}
