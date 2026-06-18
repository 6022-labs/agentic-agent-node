package interfaces

import "context"

// AgentChainRecovererInterface rebuilds the agent row from secrets + chain when the
// DB is empty: it reuses the agent id preserved in the secret store so the wallet
// path keeps resolving, then fills the identity from the on-chain mint.
type AgentChainRecovererInterface interface {
	// Recover returns recovered=true when an agent row was (re)created from a persisted wallet.
	Recover(ctx context.Context) (recovered bool, err error)
}
