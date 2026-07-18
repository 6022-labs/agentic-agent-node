package interfaces

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// ChallengeEntry is the stored challenge context needed to rebuild the exact EIP-4361 message at verify time.
type ChallengeEntry struct {
	Nonce     string
	ChainId   uint64
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// NonceStoreInterface owns single-use login challenges keyed by wallet — a new
// challenge replaces the previous one; unknown, used, or expired reads return nil.
type NonceStoreInterface interface {
	Generate(address common.Address, chainId uint64) (*ChallengeEntry, error)
	Consume(address common.Address) *ChallengeEntry
}
