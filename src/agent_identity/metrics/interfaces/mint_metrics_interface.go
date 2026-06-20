package interfaces

import "context"

// MintMetricsInterface records agent NFT mint-reconciliation outcomes.
type MintMetricsInterface interface {
	RecordMintSucceeded(ctx context.Context)
	RecordMintFailed(ctx context.Context)
}
