package interfaces

import "context"

// EnsPublisherInterface owns the domain orchestration of ENS publishing: it
// loads the agent/state, guards on mint, and decides init vs refresh-diff,
// delegating the on-chain reads/writes to EnsRegistrarInterface.
type EnsPublisherInterface interface {
	// PublishRuntimeRecords writes the agent's runtime records on chain, seeding
	// them on first run and refreshing only changed values afterwards. Returns
	// ErrIdentityNotFound when the agent is not minted yet.
	PublishRuntimeRecords(ctx context.Context) error
	// IsProvisioned reports whether the agent's ENS subnode already has records.
	// Returns ErrIdentityNotFound when the agent is not minted yet.
	IsProvisioned(ctx context.Context) (bool, error)
}
