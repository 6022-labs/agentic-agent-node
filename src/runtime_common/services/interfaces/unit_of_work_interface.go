package interfaces

import (
	"context"
	"errors"
)

// ErrUnitOfWorkRolledBack is returned by Commit when a joined scope rolled the work back.
var ErrUnitOfWorkRolledBack = errors.New("unit of work rolled back by a nested scope")

// UnitOfWorkInterface is one atomic scope, owned by a single goroutine. It tracks no changes:
// repositories called with Context() join it, and the caller decides whether it is kept.
type UnitOfWorkInterface interface {
	// Context carries the scope; every repository call inside it must use this context,
	// or it runs outside the transaction (and blocks on sqlite's single connection).
	Context() context.Context
	// Commit keeps the work; it is a no-op on a joined scope, and fails with
	// ErrUnitOfWorkRolledBack once a joined scope has rolled back.
	Commit() error
	// Rollback discards the work; on a joined scope it dooms the outer scope's Commit.
	Rollback() error
}

// UnitOfWorkFactoryInterface opens an atomic scope for any domain that needs one.
type UnitOfWorkFactoryInterface interface {
	// Start joins the open scope the context carries, else opens a new one; a context
	// whose scope already ended opens a new one too.
	Start(ctx context.Context) (UnitOfWorkInterface, error)
}
