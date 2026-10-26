package interfaces

import "context"

// UnitOfWorkInterface is one atomic scope. It tracks no changes: repositories called with
// Context() join it, and the caller decides what is written and whether it is kept.
type UnitOfWorkInterface interface {
	// Context carries the scope; repositories called with it join it.
	Context() context.Context
	// Commit and Rollback are idempotent, so `defer Rollback()` is safe after a commit.
	Commit() error
	Rollback() error
}

// UnitOfWorkFactoryInterface opens an atomic scope for any domain that needs one.
type UnitOfWorkFactoryInterface interface {
	// Start joins the scope the context already carries, if there is one.
	Start(ctx context.Context) (UnitOfWorkInterface, error)
}
